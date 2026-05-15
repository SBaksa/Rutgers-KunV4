package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

type customCommandData struct {
	Name     string `json:"name"`
	Response string `json:"response"`
	Author   string `json:"author"`
}

func CCAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!cc add <name> <response>`")
		return err
	}

	name := strings.ToLower(args[0])
	response := strings.Join(args[1:], " ")

	data := customCommandData{
		Name:     name,
		Response: response,
		Author:   m.Author.ID,
	}

	database.Instance.SetCustomCommand(m.GuildID, name, data)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Custom command `!%s` added.", name))
	return err
}

func CCRemove(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!cc remove <name>`")
		return err
	}

	name := strings.ToLower(args[0])
	database.Instance.RemoveCustomCommand(m.GuildID, name)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Custom command `!%s` removed.", name))
	return err
}

func CCList(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	commands, err := database.Instance.GetAllCustomCommands(m.GuildID)
	if err != nil || len(commands) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "No custom commands set for this server.")
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Custom Commands",
		Color:       0xCC0033,
		Description: "`!" + strings.Join(commands, "`\n`!") + "`",
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func CCDetail(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!cc detail <name>`")
		return err
	}

	name := strings.ToLower(args[0])
	var data customCommandData
	err := database.Instance.GetCustomCommand(m.GuildID, name, &data)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Custom command `!%s` not found.", name))
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title: "!" + data.Name,
		Color: 0xCC0033,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Response", Value: data.Response},
			{Name: "Added by", Value: fmt.Sprintf("<@%s>", data.Author)},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func CC(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!cc add/remove/list/detail`")
		return err
	}

	subcommand := strings.ToLower(args[0])
	rest := args[1:]

	switch subcommand {
	case "add":
		return CCAdd(s, m, rest, log, vm)
	case "remove":
		return CCRemove(s, m, rest, log, vm)
	case "list":
		return CCList(s, m, rest, log, vm)
	case "detail":
		return CCDetail(s, m, rest, log, vm)
	default:
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!cc add/remove/list/detail`")
		return err
	}
}
