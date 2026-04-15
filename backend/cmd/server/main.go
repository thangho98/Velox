package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thawng/velox/internal/config"
	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/logger"
)

const version = "velox v0.1.1"

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}
	logger.Setup()

	if handleCommand(os.Args) {
		return
	}

	runServer()
}

func handleCommand(args []string) bool {
	if len(args) <= 1 {
		return false
	}

	switch args[1] {
	case "migrate":
		runMigrate()
		return true
	case "version":
		fmt.Println(version)
		return true
	case "blurhash":
		runBlurhash()
		return true
	default:
		return false
	}
}

func runMigrate() {
	cfg := config.Load()

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	subcmd := "up"
	if len(os.Args) > 2 {
		subcmd = os.Args[2]
	}

	switch subcmd {
	case "up":
		if err := database.Migrate(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		log.Println("migrations applied successfully")

	case "rollback":
		if err := database.MigrateRollback(db); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		log.Println("rollback completed")

	case "status":
		statuses, err := database.MigrateStatus(db)
		if err != nil {
			log.Fatalf("failed to get status: %v", err)
		}

		fmt.Printf("%-8s %-30s %-10s %s\n", "VERSION", "NAME", "STATUS", "APPLIED AT")
		fmt.Println("-------- ------------------------------ ---------- -------------------")
		for _, s := range statuses {
			status := "pending"
			appliedAt := ""
			if s.Applied {
				status = "applied"
				appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("%03d      %-30s %-10s %s\n", s.Version, s.Name, status, appliedAt)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown migrate command: %s\n", subcmd)
		fmt.Fprintln(os.Stderr, "usage: velox migrate [up|rollback|status]")
		os.Exit(1)
	}
}

func runServer() {
	cfg := config.Load()

	app, err := newServerApp(cfg)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	defer app.Close()

	cleanup := app.startBackgroundServices()
	defer cleanup()

	if err := serve(app.newHTTPServer()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
