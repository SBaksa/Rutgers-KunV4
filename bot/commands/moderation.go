package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/validation"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Echo(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!echo <text>` or `!echo #channel <text>`")
		return err
	}

	targetChannelID := m.ChannelID
	message := strings.Join(args, " ")

	if strings.HasPrefix(args[0], "<#") && len(args) >= 2 {
		if !IsModerator(s, m) {
			_, err := s.ChannelMessageSend(m.ChannelID, "You don't have permission to echo to another channel.")
			return err
		}
		targetChannelID = strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(args[0], "<#"), ">"), "!")
		message = strings.Join(args[1:], " ")
	}

	_, err := s.ChannelMessageSend(targetChannelID, message)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to send message: %v", err))
		return sendErr
	}

	return nil
}

func Ignore(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
		return err
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!ignore #channel`")
		return err
	}

	channelMention := args[0]
	channelID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(channelMention, "<#"), ">"), "!")

	key := fmt.Sprintf("ignored:%s", channelID)
	var ignored bool
	err := database.Instance.GetGuildSetting(m.GuildID, key, &ignored)

	if err == nil && ignored {
		database.Instance.RemoveGuildSetting(m.GuildID, key)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully unignored <#%s>.", channelID))
	} else {
		database.Instance.SetGuildSetting(m.GuildID, key, true)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully ignored <#%s>.", channelID))
	}

	return nil
}

func ListIgnored(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
		return err
	}

	allSettings, err := database.Instance.GetAllGuildSettings(m.GuildID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "Failed to fetch settings.")
		return sendErr
	}

	var channels []string
	for key := range allSettings {
		if strings.HasPrefix(key, "ignored:") {
			channelID := strings.TrimPrefix(key, "ignored:")
			channels = append(channels, fmt.Sprintf("<#%s>", channelID))
		}
	}

	if len(channels) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "No channels are currently ignored.")
		return err
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Ignored channels:\n"+strings.Join(channels, "\n"))
	return err
}

func NetID(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if !HasManageServer(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
		return err
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!netid <netid>`")
		return err
	}

	netID := validation.NormalizeNetID(args[0])

	foundUserID, err := database.Instance.FindUserByNetID(netID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "Failed to query database.")
		return sendErr
	}
	if foundUserID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("NetID `%s` not found.", netID))
		return err
	}

	user, err := s.User(foundUserID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User found but could not fetch details: %v", err))
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{Name: fmt.Sprintf("NetID %s found", netID)},
		Title:  user.String(),
		Color:  0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: fmt.Sprintf("<@%s> (%s)", foundUserID, foundUserID)},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

// getPermissions fetches guild and member once and returns (isMod, isAdmin).
// isMod = KICK_MEMBERS or ADMINISTRATOR; isAdmin = ADMINISTRATOR only.
func getPermissions(s *discordgo.Session, m *discordgo.MessageCreate) (isMod bool, isAdmin bool) {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return false, false
	}

	if guild.OwnerID == m.Author.ID {
		return true, true
	}

	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		return false, false
	}

	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID != roleID {
				continue
			}
			if role.Permissions&discordgo.PermissionAdministrator != 0 {
				return true, true
			}
			if role.Permissions&discordgo.PermissionKickMembers != 0 {
				isMod = true
			}
		}
	}

	return isMod, false
}

func IsModerator(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	mod, _ := getPermissions(s, m)
	return mod
}

func IsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	_, admin := getPermissions(s, m)
	return admin
}

// HasManageServer returns true if the user has Manage Server or Administrator.
// This sits between Mod and Admin and gates netid-related commands.
func HasManageServer(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return false
	}
	if guild.OwnerID == m.Author.ID {
		return true
	}
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		return false
	}
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				if role.Permissions&discordgo.PermissionAdministrator != 0 ||
					role.Permissions&discordgo.PermissionManageServer != 0 {
					return true
				}
			}
		}
	}
	return false
}

func isOwner(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if m.GuildID == "" {
		return false
	}
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return false
	}
	return guild.OwnerID == m.Author.ID
}
