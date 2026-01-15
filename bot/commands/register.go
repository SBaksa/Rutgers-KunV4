package commands

import "github.com/bwmarrin/discordgo"

type CommandFunc func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)

var Registry = map[string]CommandFunc{
	"ping":     Ping,
	"roll":     Roll,
	"help":     Help,
	"course":   Course,
	"dbtest":   DBTest,
	"dbcompat": DBCompat,
	"dbdebug":  DBDebug,
	"agree": Agree,
}
