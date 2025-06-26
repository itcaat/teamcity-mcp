package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"teamcity-mcp/internal/config"
	"teamcity-mcp/internal/logging"
	"teamcity-mcp/internal/metrics"
	"teamcity-mcp/internal/server"
)

var (
	transport = flag.String("transport", "http", "Transport mode: http or stdio")
	version   = flag.Bool("version", false, "Show version information")
	envHelp   = flag.Bool("help", false, "Show environment variable help")
)

const (
	appName    = "teamcity-mcp"
	appVersion = "1.0.0"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nTeamCity MCP Server - connects TeamCity to AI agents via MCP protocol\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  Run '%s --help' for detailed environment variable documentation\n\n", os.Args[0])
	}
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", appName, appVersion)
		os.Exit(0)
	}

	if *envHelp {
		config.PrintEnvHelp()
		os.Exit(0)
	}

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logging
	logger, err := logging.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logger.Sync()

	// Initialize metrics
	metrics.Init()

	// Create server
	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create server", "error", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGHUP:
				logger.Info("Received SIGHUP, reloading configuration")
				if newCfg, err := config.Load(); err != nil {
					logger.Error("Failed to reload configuration", "error", err)
				} else {
					srv.UpdateConfig(newCfg)
				}
			case syscall.SIGINT, syscall.SIGTERM:
				logger.Info("Received shutdown signal", "signal", sig)
				cancel()
			}
		}
	}()

	// Start server
	logger.Info("Starting TeamCity MCP server",
		"version", appVersion,
		"transport", *transport,
		"teamcity_url", cfg.TeamCity.URL)

	if err := srv.Start(ctx, *transport); err != nil {
		logger.Fatal("Server failed", "error", err)
	}

	logger.Info("Server shutdown complete")
}
