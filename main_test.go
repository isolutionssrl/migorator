package main

import "testing"

// Check for valid regex pattern
func TestSplit(t *testing.T) {
	res := splitter.MatchString("SELECT 1\nGO\nSELECT 2")
	if !res {
		t.Error("Splitter failed to match")
	}
}

func TestCli(t *testing.T) {
	testCfg := Config{
		connectionString: "Server=localhost;Database=master;User Id=sa;Password=Password123;",
		migrationPath:    "./",
		runModified:      false,
		installState:     false,
		avoidTransaction: false,
	}
	if !testCfg.IsValid() {
		t.Error("Config parse failed")
	}
}

func TestCliInvalid(t *testing.T) {
	testCfg := Config{
		connectionString: "",
		migrationPath:    "./",
		runModified:      false,
		installState:     false,
		avoidTransaction: false,
	}
	if testCfg.IsValid() {
		t.Error("Config parse failed")
	}
}
