package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/umit144/kraken-portfolio/internal/api"
	"github.com/umit144/kraken-portfolio/internal/config"
	"github.com/umit144/kraken-portfolio/internal/ui"
)

type flags struct {
	envFile string
	debug   bool
}

func parseFlags() *flags {
	f := &flags{}
	flag.StringVar(&f.envFile, "env", ".env", "Path to env file")
	flag.BoolVar(&f.debug, "debug", false, "Enable debug logging")
	flag.Parse()
	return f
}

func setupLogger(debug bool) *log.Logger {
	if debug {
		return log.New(os.Stdout, "[DEBUG] ", log.LstdFlags)
	}
	return log.New(os.Stdout, "", 0)
}

func setupSignalHandler(client *api.Client, logger *log.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Printf("\nReceived signal: %v\n", sig)
		if err := client.Close(); err != nil {
			logger.Printf("Error closing client: %v\n", err)
		}
		os.Exit(0)
	}()
}

func run(f *flags, logger *log.Logger) error {
	cfg, err := config.LoadConfig(f.envFile)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	client := api.NewClient(cfg)
	if err := client.Connect(); err != nil {
		return err
	}
	defer client.Close()

	display := ui.NewDisplay()
	setupSignalHandler(client, logger)

	logger.Println("Connected to Kraken. Press Ctrl+C to exit.")
	client.StartStreaming(display.RenderPortfolio)
	return nil
}

func main() {
	flags := parseFlags()
	logger := setupLogger(flags.debug)

	if err := run(flags, logger); err != nil {
		logger.Fatalf("Error: %v\n", err)
	}
}
