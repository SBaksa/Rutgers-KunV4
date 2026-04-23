package commands

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

var eightBallResponses = []string{
	"It is certain.",
	"Without a doubt.",
	"Yes - definitely.",
	"As I see it, yes.",
	"Signs point to yes.",
	"Don't count on it.",
	"My reply is no.",
	"My sources say no.",
	"Outlook not so good.",
	"Very doubtful.",
}

func EightBall(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Ask me a question! `!8ball will I pass Calc 2?`")
		return err
	}

	question := strings.Join(args, " ")
	response := eightBallResponses[rand.Intn(len(eightBallResponses))]

	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎱 **%s**\n%s", question, response))
	return err
}

func Love(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) < 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!love <person1> <person2>`")
		return err
	}

	one := strings.ToLower(args[0])
	two := strings.ToLower(args[1])

	var percent int
	if one == two {
		percent = 100
	} else {
		percent = (calcValFromStr(one) + calcValFromStr(two)) % 100
	}

	heart := ""
	if percent == 100 {
		heart = "! ❤️"
	}

	bar := generateProgressBar(percent)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s loves %s %d%%%s\n%s", one, two, percent, heart, bar))
	return err
}

func calcValFromStr(str string) int {
	total := 0
	for _, c := range str {
		total += int(c)
	}
	return total
}

func generateProgressBar(percent int) string {
	numHashes := percent / 5
	bar := "["
	for i := 0; i < 20; i++ {
		if i < numHashes {
			bar += "#"
		} else {
			bar += " "
		}
	}
	bar += "]"
	return "`" + bar + "`"
}

func Meow(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	resp, err := http.Get("https://api.thecatapi.com/v1/images/search")
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Couldn't fetch a cat image.")
		return sendErr
	}
	defer resp.Body.Close()

	var result []struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Couldn't fetch a cat image.")
		return sendErr
	}

	_, err = s.ChannelMessageSend(m.ChannelID, result[0].URL)
	return err
}

func Woof(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	resp, err := http.Get("https://dog.ceo/api/breeds/image/random")
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Couldn't fetch a dog image.")
		return sendErr
	}
	defer resp.Body.Close()

	var result struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Status != "success" {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Couldn't fetch a dog image.")
		return sendErr
	}

	_, err = s.ChannelMessageSend(m.ChannelID, result.Message)
	return err
}
