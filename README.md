# MiGOrator

A golang migration tool for Sql Server Database.

## Installation

1. Clone repository
2. Go to project root `cd migorator`
3. Build the project `go build .`
4. Install the executable on your system `go install`

## Usage

```
dotnet-badgie-migrator <connection string> [drive:][path][filename] [-f] [-i] [-n]
  -f runs mutated migrations
  -i if needed, installs the db table needed to store state
  -n avoids wrapping each execution in a transaction
```

