package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func ReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	log.Printf("Reaction added: %s by %s on message %s", r.Emoji.Name, r.UserID, r.MessageID)
}

func MemberUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	log.Printf("Member updated: %s in guild %s", m.User.ID, m.GuildID)
}

func GuildMemberAddHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	log.Printf("Member joined: %s in guild %s", m.User.ID, m.GuildID)
}

func GuildMemberRemoveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	log.Printf("Member left: %s in guild %s", m.User.ID, m.GuildID)
}

func MessageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	log.Printf("Message deleted: %s in channel %s", m.ID, m.ChannelID)
}

func MessageUpdateHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author != nil && m.Author.Bot {
		return
	}

	log.Printf("Message edited: %s in channel %s", m.ID, m.ChannelID)
}
