package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()
	command := args[0]

	switch command {
	case "up":
		log.Println("Applying pending database migrations...")
		if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}

	case "down":
		log.Println("Rolling back last database migration...")
		if err := db.RollbackMigration(cfg.DatabaseURL); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}

	case "status", "version":
		version, dirty, err := db.GetMigrationVersion(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to fetch migration status: %v", err)
		}
		fmt.Printf("Current Migration Version: %d (Dirty: %t)\n", version, dirty)

	case "force":
		if len(args) < 2 {
			log.Fatal("Error: 'force' command requires a version argument. Usage: migrate force <version>")
		}
		ver, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("Invalid version number '%s': %v", args[1], err)
		}
		if err := db.ForceVersion(cfg.DatabaseURL, ver); err != nil {
			log.Fatalf("Migration force failed: %v", err)
		}

	default:
		log.Printf("Unknown command: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Ticket Booking Database Migration Tool

Usage:
  go run cmd/migrate/main.go <command> [arguments]

Commands:
  up                Apply all pending migrations
  down              Roll back the latest migration step
  status / version  Display current migration version and dirty status
  force <version>   Force set migration version to clear dirty state`)
}
