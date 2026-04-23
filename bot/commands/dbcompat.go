package commands

import (
	"fmt"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func DBCompat(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if database.Instance == nil {
		s.ChannelMessageSend(m.ChannelID, "Database not initialized")
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Database Compatibility Test",
		Color:       0x0099FF,
		Description: "Testing compatibility with JavaScript bot database format...",
	}

	sentMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		return err
	}

	err = database.Instance.TestJSCompatibility()

	if err != nil {
		embed.Color = 0xFF0000
		embed.Description = fmt.Sprintf("Compatibility test failed: %v", err)
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Status",
				Value: "Database format is NOT compatible with JS bot",
			},
		}
	} else {
		embed.Color = 0x00FF00
		embed.Description = "All compatibility tests passed!"
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Agreement Roles",
				Value:  "Compatible",
				Inline: true,
			},
			{
				Name:   "Guild Settings",
				Value:  "Compatible",
				Inline: true,
			},
			{
				Name:   "Global Settings",
				Value:  "Compatible",
				Inline: true,
			},
			{
				Name:   "User Data",
				Value:  "Compatible",
				Inline: true,
			},
			{
				Name:   "Custom Commands",
				Value:  "Compatible",
				Inline: true,
			},
			{
				Name:  "Status",
				Value: "Ready for seamless migration from JS bot",
			},
		}
	}

	_, err = s.ChannelMessageEditEmbed(m.ChannelID, sentMsg.ID, embed)
	return err
}

func DBDebug(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if database.Instance == nil {
		s.ChannelMessageSend(m.ChannelID, "Database not initialized")
		return nil
	}

	database.Instance.DebugDatabaseContents()

	embed := &discordgo.MessageEmbed{
		Title:       "Database Debug",
		Color:       0x0099FF,
		Description: "Database contents have been printed to console logs. Check your terminal.",
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
