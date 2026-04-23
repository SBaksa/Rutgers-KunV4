package bot

import (
	"context"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/bot/commands"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate, processor *CommandProcessor, log *logger.Logger, vm *verification.VerificationManager) {
	if m.Author.Bot {
		return
	}

	if m.GuildID == "" {
		vm.ProcessDMMessage(s, m)
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

	log.Debug("Command received", "command", command, "author", m.Author.Username, "guild", m.GuildID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := &CommandJob{
		Session: s,
		Message: m,
		Command: command,
		Args:    args,
		Ctx:     ctx,
	}

	if err := processor.Submit(job); err != nil {
		log.Error("Failed to submit command job", "command", command, "error", err)
		return
	}

	if handler, ok := commands.Registry[command]; ok {
		if err := handler(s, m, args, log, vm); err != nil {
			log.Error("Command execution failed", "command", command, "error", err)
			s.ChannelMessageSend(m.ChannelID, "❌ An error occurred while processing your command.")
		}
	} else {
		log.Debug("Unknown command", "command", command, "author", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "❓ Unknown command. Type `!help`.")
	}
}
