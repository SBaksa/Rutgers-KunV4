package bot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/bot/commands"
	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

// matches discord.com, ptb.discord.com, canary.discord.com, discordapp.com (old)
var msgLinkRegex = regexp.MustCompile(`https://(?:(?:ptb|canary)\.)?discord(?:app)?\.com/channels/(\d+)/(\d+)/(\d+)`)

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

	// Agreement channel enforcement: delete non-moderator messages that aren't !agree
	agreementChannelID, _ := database.Instance.GetGuildSettingString(m.GuildID, "agreementChannel")
	if agreementChannelID != "" && m.ChannelID == agreementChannelID && !commands.IsModerator(s, m) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		if m.Content != "!agree" {
			warning, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Please make sure you send `!agree`. You sent `%s`.", m.Content))
			if warning != nil {
				go func() {
					time.Sleep(10 * time.Second)
					s.ChannelMessageDelete(m.ChannelID, warning.ID)
				}()
			}
			return
		}
	}

	// Skip command processing in ignored channels
	var ignored bool
	if err := database.Instance.GetGuildSetting(m.GuildID, fmt.Sprintf("ignored:%s", m.ChannelID), &ignored); err == nil && ignored {
		return
	}

	commands.TrackWordCount(m.Author.ID, m.Content)

	if match := msgLinkRegex.FindStringSubmatch(m.Content); match != nil {
		handleMessageLinkPreview(s, m, match[1], match[2], match[3], log)
	}

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

	log.Debug("Command received", "command", command, "author", m.Author.Username, "guild", m.GuildID, "msg_id", m.ID, "args", args)

	job := &CommandJob{
		Session: s,
		Message: m,
		Command: command,
		Args:    args,
	}

	if err := processor.Submit(job); err != nil {
		log.Error("Failed to submit command job", "command", command, "error", err)
	}
}

func handleMessageLinkPreview(s *discordgo.Session, m *discordgo.MessageCreate, guildID, channelID, messageID string, log *logger.Logger) {
	if guildID != m.GuildID {
		return
	}

	linked, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		log.Debug("Failed to fetch linked message", "channel", channelID, "message", messageID, "error", err)
		return
	}

	if linked.Author == nil || linked.Author.Bot {
		return
	}

	content := linked.Content
	if content == "" {
		if len(linked.Attachments) > 0 {
			content = "*[Attachment]*"
		} else if len(linked.Embeds) > 0 {
			content = "*[Embed]*"
		} else {
			return
		}
	}
	if len(content) > 1024 {
		content = content[:1021] + "..."
	}

	jumpURL := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)

	guildName := guildID
	if guild, gErr := s.State.Guild(guildID); gErr == nil {
		guildName = guild.Name
	}

	msgTime := linked.Timestamp.Format("01/02/06, 3:04 PM")
	footerText := fmt.Sprintf("Requested by %s in %s•%s", m.Author.String(), guildName, msgTime)

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    linked.Author.String(),
			IconURL: linked.Author.AvatarURL("64"),
			URL:     jumpURL,
		},
		Description: content,
		Color:       0x5865F2,
		Footer:      &discordgo.MessageEmbedFooter{Text: footerText},
	}

	for _, att := range linked.Attachments {
		lower := strings.ToLower(att.Filename)
		if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") ||
			strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") ||
			strings.HasSuffix(lower, ".webp") {
			embed.Image = &discordgo.MessageEmbedImage{URL: att.URL}
			break
		}
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
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
