package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/bwmarrin/discordgo"
)

func ReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}
}

func MemberUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate, log *logger.Logger) {
	if m.User != nil && m.User.Bot {
		return
	}
}

func GuildMemberAddHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd, log *logger.Logger) {
	if m.User == nil {
		return
	}

	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		guild, err = s.Guild(m.GuildID)
		if err != nil {
			log.Error("Failed to fetch guild on member add", "guild", m.GuildID, "error", err)
			return
		}
	}

	// Log to log channel
	logChannelID, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel")
	if err == nil && logChannelID != "" {
		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    "Member joined",
				IconURL: guild.IconURL("256"),
			},
			Title: m.User.String(),
			Color: 0x00CC00,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: m.User.AvatarURL("256"),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		s.ChannelMessageSendEmbed(logChannelID, embed)
	}

	// Send welcome message
	welcomeChannelID, _ := database.Instance.GetGuildSettingString(m.GuildID, "welcomeChannel")
	welcomeText, _ := database.Instance.GetGuildSettingString(m.GuildID, "welcomeText")
	if welcomeChannelID != "" && welcomeText != "" {
		msg := strings.NewReplacer(
			"[user]", fmt.Sprintf("<@%s>", m.User.ID),
			"[guild]", guild.Name,
		).Replace(welcomeText)
		s.ChannelMessageSend(welcomeChannelID, msg)
	}
}

func GuildMemberRemoveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove, log *logger.Logger) {
	if m.User == nil {
		return
	}

	logChannelID, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel")
	if err != nil || logChannelID == "" {
		return
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Member left",
		},
		Title: m.User.String(),
		Color: 0xFF0000,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: m.User.AvatarURL("256"),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if guild, err := s.State.Guild(m.GuildID); err == nil {
		embed.Author.IconURL = guild.IconURL("256")
	}

	s.ChannelMessageSendEmbed(logChannelID, embed)
}

func MessageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete, log *logger.Logger) {
	if m.GuildID == "" {
		return
	}

	logChannelID, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel")
	if err != nil || logChannelID == "" {
		return
	}

	// Skip deletions in the agreement channel (bot cleans those up itself)
	agreementChannelID, _ := database.Instance.GetGuildSettingString(m.GuildID, "agreementChannel")
	if agreementChannelID != "" && m.ChannelID == agreementChannelID {
		return
	}

	content := "*Message content unavailable*"
	authorStr := "*Unknown*"

	if m.BeforeDelete != nil {
		if m.BeforeDelete.Author != nil {
			if m.BeforeDelete.Author.Bot {
				return
			}
			authorStr = fmt.Sprintf("<@%s> (%s)", m.BeforeDelete.Author.ID, m.BeforeDelete.Author.Username)
		}
		if m.BeforeDelete.Content != "" {
			content = m.BeforeDelete.Content
		}
	}

	if len(content) > 1024 {
		content = content[:1021] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Author:    &discordgo.MessageEmbedAuthor{Name: "Message deleted"},
		Color:     0xFF6600,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: authorStr, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", m.ChannelID), Inline: true},
			{Name: "Content", Value: content},
		},
	}

	s.ChannelMessageSendEmbed(logChannelID, embed)
}

func MessageUpdateHandler(s *discordgo.Session, m *discordgo.MessageUpdate, log *logger.Logger) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	if m.GuildID == "" || m.Content == "" {
		return
	}

	logChannelID, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel")
	if err != nil || logChannelID == "" {
		return
	}

	// Skip if nothing actually changed (embed-only update)
	if m.BeforeUpdate != nil && m.BeforeUpdate.Content == m.Content {
		return
	}

	oldContent := "*Original content unavailable*"
	if m.BeforeUpdate != nil && m.BeforeUpdate.Content != "" {
		oldContent = m.BeforeUpdate.Content
	}

	if len(oldContent) > 1024 {
		oldContent = oldContent[:1021] + "..."
	}
	newContent := m.Content
	if len(newContent) > 1024 {
		newContent = newContent[:1021] + "..."
	}

	embed := &discordgo.MessageEmbed{
		Author:    &discordgo.MessageEmbedAuthor{Name: "Message edited"},
		Title:     m.Author.String(),
		Color:     0xFFAA00,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", m.ChannelID), Inline: true},
			{Name: "Old content", Value: oldContent},
			{Name: "New content", Value: newContent},
		},
	}

	s.ChannelMessageSendEmbed(logChannelID, embed)
}
