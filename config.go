package main

import "flag"

type Config struct {
	connectionString string
	migrationPath    string
	runModified      bool
	installState     bool
	avoidTransaction bool
}

func (c *Config) ParseCommandLine() {
	flag.StringVar(&c.connectionString, "c", "", "connection string")
	flag.StringVar(&c.migrationPath, "p", "./", "path to migration files")
	flag.BoolVar(&c.runModified, "f", false, "runs modified migrations")
	flag.BoolVar(&c.installState, "i", false, "if needed, installs the db table to store state")
	flag.BoolVar(&c.avoidTransaction, "n", false, "avoids wrapping each migration in a transaction")

	flag.Parse()
}

func (c *Config) IsValid() bool {
	return c.connectionString != ""
}
