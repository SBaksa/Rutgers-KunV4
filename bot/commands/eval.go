package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/config"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
	"github.com/dop251/goja"
)

func Eval(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if !config.IsOwner(m.Author.ID) {
		_, err := s.ChannelMessageSend(m.ChannelID, "You don't have permission to use this command.")
		return err
	}

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!eval <js code>`")
		return err
	}

	code := strings.Join(args, " ")
	rt := goja.New()

	var output []string

	console := rt.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		output = append(output, strings.Join(parts, " "))
		return goja.Undefined()
	})
	rt.Set("console", console)

	rt.Set("guildID", m.GuildID)
	rt.Set("channelID", m.ChannelID)

	discord := rt.NewObject()

	discord.Set("send", func(call goja.FunctionCall) goja.Value {
		s.ChannelMessageSend(call.Argument(0).String(), call.Argument(1).String())
		return goja.Undefined()
	})
	discord.Set("deleteMessage", func(call goja.FunctionCall) goja.Value {
		s.ChannelMessageDelete(call.Argument(0).String(), call.Argument(1).String())
		return goja.Undefined()
	})
	discord.Set("createChannel", func(call goja.FunctionCall) goja.Value {
		ctype := discordgo.ChannelTypeGuildText
		if call.Argument(2).String() == "voice" {
			ctype = discordgo.ChannelTypeGuildVoice
		}
		s.GuildChannelCreate(call.Argument(0).String(), call.Argument(1).String(), ctype)
		return goja.Undefined()
	})
	discord.Set("deleteChannel", func(call goja.FunctionCall) goja.Value {
		s.ChannelDelete(call.Argument(0).String())
		return goja.Undefined()
	})
	discord.Set("kick", func(call goja.FunctionCall) goja.Value {
		s.GuildMemberDelete(call.Argument(0).String(), call.Argument(1).String())
		return goja.Undefined()
	})
	discord.Set("ban", func(call goja.FunctionCall) goja.Value {
		s.GuildBanCreate(call.Argument(0).String(), call.Argument(1).String(), 0)
		return goja.Undefined()
	})
	discord.Set("unban", func(call goja.FunctionCall) goja.Value {
		s.GuildBanDelete(call.Argument(0).String(), call.Argument(1).String())
		return goja.Undefined()
	})
	discord.Set("addRole", func(call goja.FunctionCall) goja.Value {
		s.GuildMemberRoleAdd(call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
		return goja.Undefined()
	})
	discord.Set("removeRole", func(call goja.FunctionCall) goja.Value {
		s.GuildMemberRoleRemove(call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
		return goja.Undefined()
	})

	rt.Set("discord", discord)

	timer := time.AfterFunc(5*time.Second, func() {
		rt.Interrupt("execution timeout")
	})
	defer timer.Stop()

	val, err := func() (v goja.Value, e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("panic: %v", r)
			}
		}()
		return rt.RunString(code)
	}()

	var result string
	if err != nil {
		result = fmt.Sprintf("Error: %s", err.Error())
	} else {
		lines := output
		if val != nil && val.Export() != nil {
			lines = append(lines, fmt.Sprintf("%v", val.Export()))
		}
		if len(lines) == 0 {
			result = "(no output)"
		} else {
			result = strings.Join(lines, "\n")
		}
	}

	if len(result) > 1900 {
		result = result[:1900] + "..."
	}

	_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```\n%s\n```", result))
	return sendErr
}
