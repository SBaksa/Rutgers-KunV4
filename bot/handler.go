package bot

import (
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/bot/commands"
	"github.com/bwmarrin/discordgo"
)

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
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

	if handler, ok := commands.Registry[command]; ok {
		handler(s, m, args)
	} else {
		s.ChannelMessageSend(m.ChannelID, "❓ Unknown command. Type `!help`.")
	}
}
