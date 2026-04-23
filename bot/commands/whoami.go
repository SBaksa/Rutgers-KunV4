package commands

import (
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func WhoAmI(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	embed := &discordgo.MessageEmbed{
		Title:       "I'm " + s.State.User.Username + "!",
		Color:       0xCC0033,
		Description: "I am a bot specially designed for the Rutgers community, built with Go and discordgo.",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: s.State.User.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Programmer",
				Value: "Originally written by Arjun Srivastav for the Rutgers Math Discord.\nRewritten in Go by SBaksa.",
			},
			{
				Name:  "Language",
				Value: "Go (Golang) with discordgo library",
			},
			{
				Name:  "Features",
				Value: "• Course lookup with caching\n• Verification system\n• Dice rolling\n• And more coming soon!",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Type !help for available commands",
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
