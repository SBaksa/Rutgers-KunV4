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

	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}

	if len(args) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!echo #channel <message>`")
		return err
	}

	channelMention := args[0]
	channelID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(channelMention, "<#"), ">"), "!")
	message := strings.Join(args[1:], " ")

	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to send message: %v", err))
		return sendErr
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "👍")
	return nil
}

func Ignore(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	if !IsModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
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
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Successfully unignored <#%s>.", channelID))
	} else {
		database.Instance.SetGuildSetting(m.GuildID, key, true)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Successfully ignored <#%s>.", channelID))
	}

	return nil
}

func NetID(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if !isOwner(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ This command is owner only.")
		return err
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!netid <netid>`")
		return err
	}

	netID := validation.NormalizeNetID(args[0])

	var storedNetID string
	var foundUserID string

	guilds, err := s.UserGuilds(100, "", "", false)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Failed to fetch guilds.")
		return sendErr
	}

	for _, guild := range guilds {
		members, err := s.GuildMembers(guild.ID, "", 1000)
		if err != nil {
			continue
		}
		for _, member := range members {
			err := database.Instance.GetUserData(member.User.ID, "netid", &storedNetID)
			if err == nil && storedNetID == netID {
				foundUserID = member.User.ID
				break
			}
		}
		if foundUserID != "" {
			break
		}
	}

	if foundUserID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("NetID `%s` not found.", netID))
		return err
	}

	user, err := s.User(foundUserID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ User found but could not fetch details: %v", err))
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title: user.Username + "#" + user.Discriminator,
		Color: 0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  fmt.Sprintf("NetID %s found", netID),
				Value: fmt.Sprintf("<@%s> (%s)", foundUserID, foundUserID),
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func IsModerator(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		return false
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		return false
	}

	if guild.OwnerID == m.Author.ID {
		return true
	}

	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if role.ID == roleID {
				if role.Permissions&discordgo.PermissionManageServer != 0 ||
					role.Permissions&discordgo.PermissionAdministrator != 0 ||
					role.Permissions&discordgo.PermissionKickMembers != 0 {
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
