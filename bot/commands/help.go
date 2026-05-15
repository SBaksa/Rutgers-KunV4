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
				Name:  "General",
				Value: "`!ping` – Check if the bot is alive\n`!help` – Show this message\n`!roll <NdN>` – Roll dice (e.g. `!roll 2d6`)\n`!echo <text>` – Repeat a message",
			},
			{
				Name:  "Course Info",
				Value: "`!course 198:111` – Get course info\n`!course 01:198:111:01` – Get specific section",
			},
			{
				Name:  "Verification",
				Value: "`!agree` – Begin Rutgers NetID verification\n`!cancel` – Cancel verification\n`!diagnose` – Check verification config\n`!roleswitch <role>` – Switch your verified role",
			},
			{
				Name:  "User Info",
				Value: "`!whois @user` – Info about a user\n`!whoami` – Info about the bot\n`!avatar @user` – Same as whois\n`!netid @user` – Show a user's NetID (mod only)",
			},
			{
				Name:  "Quotes",
				Value: "`!quote @user` – Save the last message from a user\n`!listquotes @user` – List saved quotes\n`!deletequote <index>` – Delete one of your quotes\n`!clearquotes` – Clear all your quotes",
			},
			{
				Name:  "Word Tracking",
				Value: "`!countword <word>` – Start tracking a word\n`!showword` – Show your word count\n`!deleteword` – Stop tracking",
			},
			{
				Name:  "Custom Commands",
				Value: "`!cc <name>` – Run a custom command\n`!cc add <name> <response>` – Add a command\n`!cc remove <name>` – Remove a command\n`!cc list` – List all custom commands\n`!cc detail <name>` – Show command details",
			},
			{
				Name:  "Fun",
				Value: "`!8ball <question>` – Ask the magic 8ball\n`!love @user` – Check compatibility\n`!meow` – Get a cat pic\n`!woof` – Get a dog pic",
			},
			{
				Name:  "Information",
				Value: "`!membercount` – Show server member count\n`!screenfetch` – Show bot system info\n`!fetchmessage <id>` – Fetch a message by ID",
			},
			{
				Name:  "Moderation",
				Value: "`!echo #channel <message>` – Send a message to a channel\n`!ignore #channel` – Toggle ignoring commands in a channel",
			},
			{
				Name:  "Config (Mod only)",
				Value: "`!setwelcomechannel <#channel>` – Set welcome channel\n`!setwelcometext <text>` – Set welcome message\n`!setlogchannel <#channel>` – Set mod log channel\n`!setagreementchannel <#channel>` – Set agreement channel\n`!setagreementroles <roles>` – Set verification roles\n`!listconfig` – Show current config",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use [user] and [guild] as placeholders in welcome text",
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
			{Name: "Set Value", Value: testValue},
			{Name: "Retrieved Value", Value: retrieved},
			{Name: "Status", Value: "Database is working correctly"},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
