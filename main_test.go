package main

import "testing"

func TestSplit(t *testing.T) {
	res := splitter.Split("SELECT 1\n\ngo\nSELECT 2\nGO\n", -1)
	if len(res) != 3 {
		t.Error("Splitter failed to match")
	}
}

func TestSplitNoMatch(t *testing.T) {
	res := splitter.Split("SELECT 1\n", -1)
	if len(res) == 0 {
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
