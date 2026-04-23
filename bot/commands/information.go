package commands

import (
	"fmt"
	"runtime"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

var startTime = time.Now()

func MemberCount(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Failed to fetch server info.")
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title: guild.Name + " Member Count",
		Color: 0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Total Members",
				Value:  fmt.Sprintf("%d", guild.MemberCount),
				Inline: true,
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func Screenfetch(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(startTime)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	guilds := s.State.Guilds

	embed := &discordgo.MessageEmbed{
		Title: s.State.User.Username + " System Info",
		Color: 0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Language",
				Value:  "Go " + runtime.Version(),
				Inline: true,
			},
			{
				Name:   "OS",
				Value:  runtime.GOOS + "/" + runtime.GOARCH,
				Inline: true,
			},
			{
				Name:   "Uptime",
				Value:  fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds),
				Inline: true,
			},
			{
				Name:   "Servers",
				Value:  fmt.Sprintf("%d", len(guilds)),
				Inline: true,
			},
			{
				Name:   "Memory Usage",
				Value:  fmt.Sprintf("%.2f MB", float64(memStats.Alloc)/1024/1024),
				Inline: true,
			},
			{
				Name:   "Goroutines",
				Value:  fmt.Sprintf("%d", runtime.NumGoroutine()),
				Inline: true,
			},
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func FetchMessage(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!fetchmessage <messageID>`")
		return err
	}

	msg, err := s.ChannelMessage(m.ChannelID, args[0])
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Message not found.")
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title: "Message by " + msg.Author.Username,
		Color: 0x0099FF,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: msg.Author.AvatarURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Content",
				Value: msg.Content,
			},
			{
				Name:   "Message ID",
				Value:  msg.ID,
				Inline: true,
			},
			{
				Name:   "Author ID",
				Value:  msg.Author.ID,
				Inline: true,
			},
		},
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
