# MiGOrator

A golang migration tool for Sql Server Database.

## How it works

MiGOrator takes a connection string (-c flag) and a folder path (-p) as input.
It will then look for all files in the folder path that end with `.sql` and run them in alphabetical order.

In order to keep track of which files have been run, MiGOrator will create a table called `MigoratorRuns` in the database. (If the flag -i is set)

## Installation

1. Clone repository
2. Go to project root `cd migorator`
3. Build the project `go build .`
4. Install the executable on your system `go install`

## Usage

```
dotnet-badgie-migrator -c <connection string> -p <directory path> [-f] [-i] [-n]
  -f runs mutated migrations
  -i if needed, installs the db table needed to store state
  -n avoids wrapping each execution in a transaction
```

