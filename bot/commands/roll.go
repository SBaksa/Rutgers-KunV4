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

	count := 1
	sides := 0

	if idx := strings.Index(input, "d"); idx == -1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Invalid format. Try `!roll 2d6` or `!roll d20`.")
		return err
	} else if idx == 0 {
		sides, _ = strconv.Atoi(input[1:])
	} else {
		count, _ = strconv.Atoi(input[:idx])
		sides, _ = strconv.Atoi(input[idx+1:])
	}

	if sides <= 1 || count < 1 || count > 100 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Invalid dice. Use format `2d6` — max 100 dice, min 2 sides.")
		return err
	}

	rolls := make([]string, count)
	total := 0
	for i := range rolls {
		r := rand.Intn(sides) + 1
		total += r
		rolls[i] = strconv.Itoa(r)
	}

	desc := fmt.Sprintf("**Total: %d**", total)
	if count > 1 {
		desc = fmt.Sprintf("Rolls: %s\n%s", strings.Join(rolls, ", "), desc)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%dd%d", count, sides),
		Color:       0x0099FF,
		Description: desc,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    fmt.Sprintf("Rolled by %s", m.Author.Username),
			IconURL: m.Author.AvatarURL(""),
		},
	}

	_, embedErr := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return embedErr
}