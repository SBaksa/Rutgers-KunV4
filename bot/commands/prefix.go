package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Prefix(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	current := "!"
	if p, err := database.Instance.GetGuildSettingString(m.GuildID, "prefix"); err == nil && p != "" {
		current = p
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current prefix: `%s`", current))
		return err
	}

	if !isOwner(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "Only the server owner can change the prefix.")
		return err
	}

	newPrefix := args[0]

	if newPrefix == "reset" {
		if err := database.Instance.RemoveGuildSetting(m.GuildID, "prefix"); err != nil {
			_, sendErr := s.ChannelMessageSend(m.ChannelID, "Failed to reset prefix.")
			return sendErr
		}
		_, err := s.ChannelMessageSend(m.ChannelID, "Prefix reset to `!`.")
		return err
	}

	if strings.ContainsAny(newPrefix, " \t\n") || len(newPrefix) > 5 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Prefix must be 1–5 characters with no spaces.")
		return err
	}

	if err := database.Instance.SetGuildSetting(m.GuildID, "prefix", newPrefix); err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "Failed to update prefix.")
		return sendErr
	}

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Prefix updated to `%s`.", newPrefix))
	return err
}
