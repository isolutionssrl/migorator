package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/microsoft/go-mssqldb"
)

var (
	cfg = Config{}
)

type MigrationResult int

const (
	Success  MigrationResult = 0
	Failed   MigrationResult = 1
	Modified MigrationResult = 2
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
	for _, file := range files {
		sql := readFileContent(file)

		// Check if file has already been run
		hasher := md5.New()
		hasher.Write(sql)
		hash := hex.EncodeToString(hasher.Sum(nil))
		runHash := getHashIfRunned(db, file)
		if runHash != "" {
			if hash == runHash {
				fmt.Printf("Skipped - %s\n", file)
				continue
			}

			if !cfg.runModified {
				runFile(db, string(sql))
				db.Exec("UPDATE [dbo].[MigoratorRuns] SET LastRun = GETDATE(), MD5 = ?, Result = ? WHERE FileName = ?", hash, Modified, filepath.Base(file))
				fmt.Printf("Modified - %s\n", file)
				continue
			}
		}

		runFile(db, string(sql))
		db.Exec("INSERT INTO [dbo].[MigoratorRuns] (FileName, LastRun, MD5, Result) VALUES (?, GETDATE(), ?, ?)", filepath.Base(file), hash, Success)
		fmt.Printf("Run - %s\n", file)
	}
}

func readFileContent(path string) []byte {
	sql, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Error reading file:", err.Error())
	}

	return sql
}

func getHashIfRunned(db *sql.DB, file string) string {
	fileName := filepath.Base(file)

	query := "SELECT MD5 FROM [dbo].[MigoratorRuns] WHERE FileName = ?"

	var hash string
	err := db.QueryRow(query, fileName).Scan(&hash)
	if err == sql.ErrNoRows {
		return ""
	}

	return hash
}

func runFile(db *sql.DB, sql string) {
	if cfg.avoidTransaction {
		// Run migration without transaction
		_, err := db.Exec(string(sql))
		if err != nil {
			fmt.Println("Error running migration: ", err.Error())
		}
	} else {
		// Run migration with transaction
		tx, err := db.Begin()
		if err != nil {
			log.Fatal("Error starting transaction: ", err.Error())
		}

		_, err = tx.Exec(string(sql))
		if err != nil {
			fmt.Println("Error running migration: ", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
}

func createStateTable(db *sql.DB) {
	// Create state table if it does not exist
	command := `
	IF NOT EXISTS (SELECT * FROM sys.objects WHERE object_id = OBJECT_ID(N'[dbo].[MigoratorRuns]') AND type in (N'U'))
	BEGIN
		CREATE TABLE [dbo].[MigoratorRuns] (
			[Id] [int] IDENTITY(1,1) NOT NULL,
			[FileName] [nvarchar](max) NOT NULL,
			[LastRun] [datetime] NOT NULL,
			[MD5] [nvarchar](max) NOT NULL,
			[Result] [bit] NOT NULL,
		) ON [PRIMARY]
	END
	`
	_, err := db.Exec(command)
	if err != nil {
		log.Fatal("Error creating state table: ", err.Error())
	}
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
