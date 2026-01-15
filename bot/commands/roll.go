package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func Roll(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	rand.Seed(time.Now().UnixNano())

	if len(args) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       "Dice Roll",
			Color:       0xFF0000,
			Description: "Usage: `!roll d20`, `!roll d6`, etc.",
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	input := strings.ToLower(args[0])
	if !strings.HasPrefix(input, "d") {
		embed := &discordgo.MessageEmbed{
			Title:       "Invalid Format",
			Color:       0xFF0000,
			Description: "Try `!roll d20`",
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	sides, err := strconv.Atoi(strings.TrimPrefix(input, "d"))
	if err != nil || sides <= 1 {
		embed := &discordgo.MessageEmbed{
			Title:       "Invalid Dice",
			Color:       0xFF0000,
			Description: "Number of sides must be greater than 1",
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
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

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}
