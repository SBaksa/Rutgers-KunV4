package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

type wordCountData struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func CountWord(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!countword <word>`")
		return err
	}

	word := strings.ToLower(args[0])

	var data wordCountData
	database.Instance.GetWordCount(m.Author.ID, &data)

	data.Word = word
	data.Count = 0

	database.Instance.SetWordCount(m.Author.ID, data)

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Now counting occurrences of `%s` in your messages.", word))
	return err
}

func ShowWord(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	var data wordCountData
	err := database.Instance.GetWordCount(m.Author.ID, &data)
	if err != nil || data.Word == "" {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "You're not tracking any word. Use `!countword <word>` to start.")
		return sendErr
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You have said `%s` **%d** time(s).", data.Word, data.Count))
	return err
}

func DeleteWord(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	database.Instance.RemoveUserData(m.Author.ID, "countword")
	_, err := s.ChannelMessageSend(m.ChannelID, "Word counter deleted.")
	return err
}
