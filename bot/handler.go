package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/bot/commands"
	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate, processor *CommandProcessor, log *logger.Logger, vm *verification.VerificationManager) {
	if m.Author.Bot {
		return
	}

	if m.GuildID == "" {
		vm.ProcessDMMessage(s, m)
		return
	}

	if found, slur := containsSlur(m.Content); found {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		logSlur(s, m, slur, log)
		return
	}

	commands.TrackWordCount(m.Author.ID, m.Content)

	const prefix = "!"
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}

	content := strings.TrimPrefix(m.Content, prefix)
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	log.Debug("Command received", "command", command, "author", m.Author.Username, "guild", m.GuildID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := &CommandJob{
		Session: s,
		Message: m,
		Command: command,
		Args:    args,
		Ctx:     ctx,
	}

	if err := processor.Submit(job); err != nil {
		log.Error("Failed to submit command job", "command", command, "error", err)
	}
}

func logSlur(s *discordgo.Session, m *discordgo.MessageCreate, slur string, log *logger.Logger) {
	log.Info("Slur detected and deleted", "user", m.Author.Username, "guild", m.GuildID, "slur", slur)

	logChannelID, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel")
	if err != nil || logChannelID == "" {
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "🚫 Slur Detected",
		Color: 0xFF0000,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: fmt.Sprintf("<@%s> (%s)", m.Author.ID, m.Author.Username), Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s> (#%s)", m.ChannelID, channel.Name), Inline: true},
			{Name: "Word", Value: fmt.Sprintf("`%s`", slur), Inline: true},
		},
	}

	s.ChannelMessageSendEmbed(logChannelID, embed)
}
