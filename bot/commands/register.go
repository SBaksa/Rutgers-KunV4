package commands

import (
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

// CommandFunc defines the signature for command functions
type CommandFunc func(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error

// Registry maps command names to their handlers
var Registry = map[string]CommandFunc{
	"ping":     Ping,
	"roll":     Roll,
	"help":     Help,
	"course":   Course,
	"dbtest":   DBTest,
	"dbcompat": DBCompat,
	"dbdebug":  DBDebug,
	"agree":    Agree,
	"cancel":   Cancel,
	"whoami":   WhoAmI,
	"whois":    WhoIs,
	"who":      WhoIs,
	"avatar":   WhoIs,
	"av":       WhoIs,
	"echo":                 Echo,
	"ignore":               Ignore,
	"unignore":             Ignore,
	"netid":                NetID,
	"8ball":                EightBall,
	"eightball":            EightBall,
	"love":                 Love,
	"meow":                 Meow,
	"cat":                  Meow,
	"kitty":                Meow,
	"woof":                 Woof,
	"bark":                 Woof,
	"dog":                  Woof,
	"setwelcomechannel":    SetWelcomeChannel,
	"setwelcometext":       SetWelcomeText,
	"setlogchannel":        SetLogChannel,
	"setagreementchannel":  SetAgreementChannel,
	"setagreementroles":    SetAgreementRoles,
	"listconfig":           ListConfig,
	"listconfigs":          ListConfig,
	"configs":              ListConfig,
	"quote":                Quote,
	"addquote":             Quote,
	"listquotes":           ListQuotes,
	"listquote":            ListQuotes,
	"quotes":               ListQuotes,
	"deletequote":          DeleteQuote,
	"clearquotes":          ClearQuotes,
	"membercount":          MemberCount,
	"screenfetch":          Screenfetch,
	"fetchmessage":         FetchMessage,
	"countword":            CountWord,
	"showword":             ShowWord,
	"deleteword":           DeleteWord,
	"cc":                   CC,
	"diagnose":             Diagnose,
	"roleswitch":           RoleSwitch,
}
