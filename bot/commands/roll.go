package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Roll(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	rand.Seed(time.Now().UnixNano())

	if len(args) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "Dice Roll",
			Color:       0xFF0000,
			Description: "Usage: `!roll d20`, `!roll d6`, etc.",
		}
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return err
	}

	input := strings.ToLower(args[0])
	if !strings.HasPrefix(input, "d") {
		embed := &discordgo.MessageEmbed{
			Title:       "Invalid Format",
			Color:       0xFF0000,
			Description: "Try `!roll d20`",
		}
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return err
	}

	sides, err := strconv.Atoi(strings.TrimPrefix(input, "d"))
	if err != nil || sides <= 1 {
		embed := &discordgo.MessageEmbed{
			Title:       "Invalid Dice",
			Color:       0xFF0000,
			Description: "Number of sides must be greater than 1",
		}
		_, embedErr := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return embedErr
	}

	result := rand.Intn(sides) + 1

	embed := &discordgo.MessageEmbed{
		Title: "Dice Roll Results",
		Color: 0x0099FF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Die Type",
				Value:  fmt.Sprintf("d%d", sides),
				Inline: true,
			},
			{
				Name:   "Result",
				Value:  fmt.Sprintf("**%d**", result),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Rolled by %s", m.Author.Username),
			IconURL: m.Author.AvatarURL(""),
		},
	}

	_, embedErr := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return embedErr
}