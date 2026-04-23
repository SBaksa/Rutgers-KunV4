package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Quote(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!quote @user`")
		return err
	}

	userMention := args[0]
	userID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(userMention, "<@"), ">"), "!")

	if userID == m.Author.ID {
		_, err := s.ChannelMessageSend(m.ChannelID, "You can't quote yourself.")
		return err
	}

	user, err := s.User(userID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ User not found.")
		return sendErr
	}

	if user.Bot {
		_, err := s.ChannelMessageSend(m.ChannelID, "You can't quote a bot.")
		return err
	}

	messages, err := s.ChannelMessages(m.ChannelID, 100, m.ID, "", "")
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Couldn't fetch messages.")
		return sendErr
	}

	var lastMessage *discordgo.Message
	for _, msg := range messages {
		if msg.Author.ID == userID {
			lastMessage = msg
			break
		}
	}

	if lastMessage == nil {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s's last message could not be found in this channel.", user.Username))
		return err
	}

	if len(lastMessage.Content) > 1024 {
		_, err := s.ChannelMessageSend(m.ChannelID, "That quote is too long!")
		return err
	}

	newQuote := lastMessage.Content
	if newQuote == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Unable to save quote (empty quote).")
		return err
	}

	quotes, _ := database.Instance.GetUserQuotes(userID)

	for len(quotes) >= 25 {
		quotes = quotes[1:]
	}

	quotes = append(quotes, newQuote)
	database.Instance.SetUserQuotes(userID, quotes)

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully saved quote for user @%s!", user.Username))
	return err
}

func ListQuotes(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	targetUser := m.Author
	if len(args) > 0 {
		userID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(args[0], "<@"), ">"), "!")
		user, err := s.User(userID)
		if err == nil {
			targetUser = user
		}
	}

	quotes, err := database.Instance.GetUserQuotes(targetUser.ID)
	if err != nil || len(quotes) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "This user has no quotes. :(")
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title: "Quotes for " + targetUser.Username,
		Color: 0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: targetUser.AvatarURL("256"),
		},
	}

	start := 0
	if len(quotes) > 5 {
		start = len(quotes) - 5
	}

	for i := start; i < len(quotes); i++ {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("Quote %d", i+1),
			Value: quotes[i],
		})
	}

	if len(quotes) > 5 {
		embed.Description = fmt.Sprintf("Showing last 5 of %d quotes.", len(quotes))

		dmChannel, dmErr := s.UserChannelCreate(m.Author.ID)
		if dmErr == nil {
			fullEmbed := &discordgo.MessageEmbed{
				Title: "All Quotes for " + targetUser.Username,
				Color: 0xCC0033,
			}
			for i, q := range quotes {
				fullEmbed.Fields = append(fullEmbed.Fields, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("Quote %d", i+1),
					Value: q,
				})
			}
			s.ChannelMessageSendEmbed(dmChannel.ID, fullEmbed)
		}
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func DeleteQuote(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!deletequote <index>` (use `!listquotes` to see indices)")
		return err
	}

	quotes, err := database.Instance.GetUserQuotes(m.Author.ID)
	if err != nil || len(quotes) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "You have no quotes to delete.")
		return sendErr
	}

	var deleted []int
	for _, arg := range args {
		idx, err := strconv.Atoi(arg)
		if err != nil || idx < 1 || idx > len(quotes) {
			continue
		}
		deleted = append(deleted, idx-1)
	}

	if len(deleted) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ No valid indices provided.")
		return err
	}

	deleteSet := make(map[int]bool)
	for _, idx := range deleted {
		deleteSet[idx] = true
	}

	var newQuotes []string
	for i, q := range quotes {
		if !deleteSet[i] {
			newQuotes = append(newQuotes, q)
		}
	}

	database.Instance.SetUserQuotes(m.Author.ID, newQuotes)

	deletedStrs := make([]string, len(deleted))
	for i, idx := range deleted {
		deletedStrs[i] = strconv.Itoa(idx + 1)
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully deleted quote(s) %s.", strings.Join(deletedStrs, ", ")))
	return err
}

func ClearQuotes(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	database.Instance.SetUserQuotes(m.Author.ID, []string{})
	_, err := s.ChannelMessageSend(m.ChannelID, "All your quotes have been cleared.")
	return err
}
