package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"telegrambot/internal/bot"
	"telegrambot/internal/config"
	"telegrambot/internal/storage"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Set up logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	store, err := storage.NewFileStorage("data")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize and start bot
	telegramBot, err := bot.New(cfg, store)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Start bot in goroutine
	go func() {
		if err := telegramBot.Start(ctx); err != nil {
			log.Errorf("Bot error: %v", err)
			cancel()
		}
	}()

	log.Info("Bot started successfully. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		log.Info("Context cancelled, shutting down...")
	}

	cancel()
	telegramBot.Stop()
	log.Info("Bot stopped.")
}
