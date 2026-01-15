package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SBaksa/Rutgers-KunV4/bot"
	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set")
	}

	// Initialize database
	if err := database.Initialize("settings.sqlite3"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Add more intents for Phase 1
	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuilds

	dg.AddHandler(bot.MessageHandler)
	// TODO: Add more handlers as we implement them
	// dg.AddHandler(bot.ReactionHandler)
	// dg.AddHandler(bot.MemberUpdateHandler)
	// dg.AddHandler(bot.GuildMemberAddHandler)
	// dg.AddHandler(bot.GuildMemberRemoveHandler)
	// dg.AddHandler(bot.MessageDeleteHandler)
	// dg.AddHandler(bot.MessageUpdateHandler)

	if err := dg.Open(); err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	log.Println("Rutgers-KunV4 is live.")

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
	if database.Instance != nil {
		database.Instance.Close()
	}
	dg.Close()
}
