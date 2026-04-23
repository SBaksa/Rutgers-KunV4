package commands

import (
	"fmt"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Help(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	embed := &discordgo.MessageEmbed{
		Title: "Rutgers-KunV4 Commands",
		Color: 0xCC0033,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Basic Commands",
				Value: "`!ping` – Check if the bot is alive\n`!help` – Show this help message",
			},
			{
				Name:  "Fun Commands",
				Value: "`!roll d20` – Roll a die with N sides",
			},
			{
				Name:  "Course Information",
				Value: "`!course 198:111` – Get course info\n`!course 01:198:111:01` – Get specific section info",
			},
			{
				Name:  "User Information",
				Value: "`!whoami` – Information about the bot\n`!whois @user` – Information about a user\n`!avatar @user` – Same as whois",
			},
			{
				Name:  "Database Commands",
				Value: "`!dbtest` – Test database functionality\n`!dbcompat` – Test JS bot compatibility\n`!dbdebug` – Show database contents",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "More features coming soon",
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func DBTest(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if database.Instance == nil {
		s.ChannelMessageSend(m.ChannelID, "Database not initialized")
		return nil
	}

	guildID := ""
	if m.GuildID != "" {
		guildID = m.GuildID
	}

	testKey := "test_setting"
	testValue := "Hello from database!"

	err := database.Instance.SetGuildSetting(guildID, testKey, testValue)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to set setting: %v", err))
		return nil
	}

	var retrieved string
	err = database.Instance.GetGuildSetting(guildID, testKey, &retrieved)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to get setting: %v", err))
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title: "Database Test Results",
		Color: 0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Set Value",
				Value: testValue,
			},
			{
				Name:  "Retrieved Value",
				Value: retrieved,
			},
			{
				Name:  "Status",
				Value: "Database is working correctly",
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
