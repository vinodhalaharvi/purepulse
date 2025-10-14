package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/vinodhalaharvi/purepulse/db"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Create database config
	cfg, err := db.NewConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to create database config: %v", err)
	}

	// Connect to database
	conn, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Create migrator
	migrator, err := db.NewMigrator(conn.DB, cfg)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer migrator.Close()

	// Execute command
	switch command {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("✓ All migrations applied successfully")

	case "down":
		if err := migrator.Down(); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("✓ Last migration rolled back successfully")

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if dirty {
			log.Printf("Current version: %d (dirty - migration failed)", version)
		} else {
			log.Printf("Current version: %d", version)
		}

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := migrator.Force(version); err != nil {
			log.Fatalf("Force migration failed: %v", err)
		}
		log.Printf("✓ Forced migration version to %d", version)

	case "steps":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate steps <n>")
		}
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid steps number: %v", err)
		}
		if err := migrator.Steps(n); err != nil {
			log.Fatalf("Migration steps failed: %v", err)
		}
		log.Printf("✓ Applied %d migration steps", n)

	case "health":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		health := conn.Health(ctx)
		fmt.Printf("Status: %s\n", health.Status)
		fmt.Printf("Latency: %v\n", health.Latency)
		fmt.Printf("Open Connections: %d/%d\n", health.Stats.OpenConns, health.Stats.MaxOpenConns)
		fmt.Printf("In Use: %d\n", health.Stats.InUse)
		fmt.Printf("Idle: %d\n", health.Stats.Idle)
		if health.Error != "" {
			fmt.Printf("Error: %s\n", health.Error)
		}

	default:
		log.Printf("Unknown command: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up             Run all pending migrations")
	fmt.Println("  down           Rollback last migration")
	fmt.Println("  version        Show current migration version")
	fmt.Println("  force <n>      Force migration version to n (use with caution)")
	fmt.Println("  steps <n>      Run n migration steps (positive for up, negative for down)")
	fmt.Println("  health         Check database health")
}
