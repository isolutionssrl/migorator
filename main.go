package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	_ "github.com/microsoft/go-mssqldb"
)

type MigrationResult int

const (
	Success  MigrationResult = 0
	Failed   MigrationResult = 1
	Modified MigrationResult = 2
)

var (
	cfg      = Config{}
	splitter = regexp.MustCompile(`(?im)\nGO\s?\n`)
)

func main() {

	cfg.ParseCommandLine()

	if !cfg.IsValid() {
		printUsage()
		os.Exit(1)
	}

	connectionString := cfg.connectionString
	migrationPath := cfg.migrationPath

	// Open database connection
	db, err := sql.Open("mssql", connectionString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer db.Close()

	// Create state table if needed
	if cfg.installState {
		createStateTable(db)
	}

	// Read migrations from directory
	files := readDirectory(migrationPath)

	// Run migrations
	runMigrations(db, files)
}

func runMigrations(db *sql.DB, files []string) {
	stateTableExists := stateTableExists(db)
	for _, file := range files {
		sql := readFileContent(file)

		// Check if file has already been run
		hasher := md5.New()
		hasher.Write([]byte(sql))
		hash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

		runHash := getHashIfRunned(db, file)

		if runHash != "" {
			if hash == runHash {
				fmt.Printf("Skipped - %s\n", file)
				continue
			}

			if cfg.runModified {
				runFile(db, string(sql), file)
				if stateTableExists {
					db.Exec("UPDATE [dbo].[MigoratorRuns] SET LastRun = GETDATE(), MD5 = @p1, MigrationResult = @p2 WHERE FileName = @p3", hash, Modified, filepath.Base(file))
				}
				fmt.Printf("Modified - %s\n", file)
				continue
			}
			log.Fatalf("Modified - %s\n", file)
		}

		fmt.Printf("Run - %s\n", file)
		runFile(db, string(sql), file)
		if stateTableExists {
			db.Exec("INSERT INTO [dbo].[MigoratorRuns] (FileName, LastRun, MD5, MigrationResult) VALUES (@p1, GETDATE(), @p2, @p3)", filepath.Base(file), hash, Success)
		}
	}
}

func readFileContent(path string) string {
	lines, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Error reading file:", err.Error())
	}

	return string(lines)
}

func getHashIfRunned(db *sql.DB, file string) string {
	fileName := filepath.Base(file)

	query := "SELECT MD5 FROM [dbo].[MigoratorRuns] WHERE FileName = ?"

	var hash string
	err := db.QueryRow(query, fileName).Scan(&hash)
	if err != nil {
		return ""
	}

	return hash
}

func runFile(db *sql.DB, sql string, file string) {
	commands := splitter.Split(sql, -1)
	for _, command := range commands {
		if command == "" {
			continue
		}

		if cfg.avoidTransaction {
			// Run migration without transaction
			_, err := db.Exec(command)
			if err != nil {
				log.Fatalf("Error running migration - %s - %s ", file, err.Error())
			}
		} else {
			// Run migration with transaction
			tx, _ := db.Begin()

			_, err := tx.Exec(command)
			if err != nil {
				tx.Rollback()
				log.Fatalf("Error running migration - %s - %s ", file, err.Error())
			} else {
				tx.Commit()
			}
		}
	}
}

func createStateTable(db *sql.DB) {
	// Create state table if it does not exist

	if stateTableExists(db) {
		return
	}

	command := `
		CREATE TABLE [dbo].[MigoratorRuns] (
			Id              INT             IDENTITY (1, 1) NOT NULL,
			LastRun         DATETIME        NOT NULL,
			Filename        NVARCHAR(2000)  NOT NULL,
			MD5             VARCHAR(50)     NOT NULL,
			MigrationResult TINYINT         NOT NULL,
			CONSTRAINT [PK_MigoratorRuns] PRIMARY KEY CLUSTERED ([Id] ASC)
		);
	`
	_, err := db.Exec(command)
	if err != nil {
		log.Fatal("Error creating state table: ", err.Error())
	}
}

func stateTableExists(db *sql.DB) bool {
	query := "SELECT COUNT(*) FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[MigoratorRuns]') AND type in (N'U')"

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Fatal("Error checking state table: ", err.Error())
	}

	return count != 0
}

func readDirectory(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	var fileNames []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			fullPath := filepath.Join(path, file.Name())
			fileNames = append(fileNames, fullPath)
		}
	}

	// Make sure the files are sorted by name
	sort.Strings(fileNames)

	return fileNames
}

func printUsage() {
	fmt.Println("Usage: migorator -c <connection string> -p <path to migration files> [-f] [-i] [-n]")
	fmt.Println("\tf: Runs mutated migrations")
	fmt.Println("\ti: If needed, installs the db table to store state")
	fmt.Println("\tn: Avoids wrapping each migration in a transaction")
}
