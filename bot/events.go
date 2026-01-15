package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// ReactionHandler handles message reaction events
func ReactionHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore bot reactions
	if r.UserID == s.State.User.ID {
		return
	}

	log.Printf("Reaction added: %s by %s on message %s", r.Emoji.Name, r.UserID, r.MessageID)

	// TODO: Implement reaction-based features:
	// - Approval system (👍/👎)
	// - Course prerequisite browsing (🇦🇧🇨...)
	// - Quote reactions (🗑️ for delete)
	// - Agreement slim reactions
}

// MemberUpdateHandler handles guild member updates (role changes, etc.)
func MemberUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	log.Printf("Member updated: %s in guild %s", m.User.ID, m.GuildID)

	// TODO: Implement member update features:
	// - Role response messages
	// - Welcome messages for new verified members
	// - Logging role changes
}

// GuildMemberAddHandler handles new members joining
func GuildMemberAddHandler(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	log.Printf("Member joined: %s in guild %s", m.User.ID, m.GuildID)

	// TODO: Implement join features:
	// - Welcome messages
	// - Auto-role assignment
	// - Logging member joins
}

// GuildMemberRemoveHandler handles members leaving
func GuildMemberRemoveHandler(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	log.Printf("Member left: %s in guild %s", m.User.ID, m.GuildID)

	// TODO: Implement leave features:
	// - Logging member leaves
	// - Cleanup of user data
}

// MessageDeleteHandler handles message deletions
func MessageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	log.Printf("Message deleted: %s in channel %s", m.ID, m.ChannelID)

	// TODO: Implement deletion features:
	// - Log deleted messages
	// - Handle chain breaking
}

// MessageUpdateHandler handles message edits
func MessageUpdateHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {
	// Ignore bot messages and messages without content changes
	if m.Author != nil && m.Author.Bot {
		return
	}

	log.Printf("Message edited: %s in channel %s", m.ID, m.ChannelID)

	// TODO: Implement edit features:
	// - Log message edits
	// - Update LaTeX renderings
}
