package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/bot"
	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	log := logger.New(logger.Info, true)
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set")
	}

	if err := database.Initialize("settings.sqlite3"); err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}
	defer func() {
		if database.Instance != nil {
			database.Instance.Close()
			log.Info("Database connection closed")
		}
	}()

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session", "error", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuilds |
		discordgo.IntentsDirectMessages

	processor := bot.NewCommandProcessor(4, log)
	processor.Start()

	verificationManager := verification.NewVerificationManager(log)

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		bot.MessageHandler(s, m, processor, log, verificationManager)
	})

	stopCleanup := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				verificationManager.CleanupExpiredVerifications()
			case <-stopCleanup:
				return
			}
		}
	}()

	if err := dg.Open(); err != nil {
		log.Fatal("Error opening connection", "error", err)
	}
	defer dg.Close()

	log.Info("Rutgers-KunV4 is live")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info("Shutdown signal received, gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	close(stopCleanup)

	log.Info("Closing Discord connection")
	dg.Close()

	processor.Shutdown()

	<-ctx.Done()
	log.Info("Graceful shutdown complete")
}
