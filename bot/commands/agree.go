package commands

import (
	"fmt"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Agree(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "This command only works in servers.")
		return err
	}

	log.Info("Agree command invoked", "user", m.Author.ID, "guild", m.GuildID)

	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err != nil || len(agreementRoles) == 0 {
		log.Debug("No agreement roles configured", "guild", m.GuildID, "error", err)
		embed := &discordgo.MessageEmbed{
			Title:       "No Agreement Configured",
			Color:       0xFF0000,
			Description: "This server has not configured the agreement system yet.",
		}
		_, embedErr := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return embedErr
	}

	agreementChannel, channelErr := database.Instance.GetAgreementChannel(m.GuildID)
	if channelErr == nil && agreementChannel != "" && agreementChannel != m.ChannelID {
		embed := &discordgo.MessageEmbed{
			Title:       "Wrong Channel",
			Color:       0xFF9900,
			Description: fmt.Sprintf("Please use <#%s> to start the agreement process.", agreementChannel),
		}
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return err
	}

	dmChannel, dmErr := s.UserChannelCreate(m.Author.ID)
	if dmErr != nil {
		log.Warn("Failed to create DM channel", "user", m.Author.ID, "error", dmErr)
		_, err := s.ChannelMessageSend(m.ChannelID, "I couldn't open a DM with you. Please make sure DMs are enabled.")
		return err
	}

	state, stateErr := vm.StartVerification(m.Author.ID, m.GuildID)
	if stateErr != nil {
		log.Error("Failed to start verification", "user", m.Author.ID, "error", stateErr)
		_, err := s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try again.")
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Welcome to the Agreement Process",
		Color:       0xCC0033,
		Description: "I'll help you verify your identity and get access to the server.\n\nPlease start by choosing which role you'd like to apply for:",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "This verification will expire in 15 minutes",
		},
	}

	roleList := ""
	roleCount := 1
	for _, role := range agreementRoles {
		if role.Authenticate == "permission" {
			continue
		}

		discordRole, discordErr := s.State.Role(m.GuildID, role.RoleID)
		if discordErr == nil {
			roleList += fmt.Sprintf("%d. %s\n", roleCount, discordRole.Name)
			roleCount++
		}
	}

	if roleList != "" {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Available Roles",
				Value: roleList,
			},
		}
	}

	_, dmSendErr := s.ChannelMessageSendEmbed(dmChannel.ID, embed)
	if dmSendErr != nil {
		log.Error("Failed to send DM", "user", m.Author.ID, "error", dmSendErr)
		_ = vm.CancelVerification(m.Author.ID)
		return dmSendErr
	}

	log.Debug("Verification initiated", "user", m.Author.ID, "guild", m.GuildID, "state", state.Step)

	confirmEmbed := &discordgo.MessageEmbed{
		Title:       "✅ Check Your DMs",
		Color:       0x00FF00,
		Description: "I've sent you a direct message to start the verification process.",
	}
	_, confirmErr := s.ChannelMessageSendEmbed(m.ChannelID, confirmEmbed)
	return confirmErr
}

func Cancel(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID != "" {
		return nil
	}

	log.Info("Cancel command invoked", "user", m.Author.ID)

	cancelErr := vm.CancelVerification(m.Author.ID)
	if cancelErr != nil {
		log.Error("Failed to cancel verification", "user", m.Author.ID, "error", cancelErr)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Verification Cancelled",
		Color:       0xFF0000,
		Description: "Your verification process has been cancelled. You can start over with `!agree` in any server.",
	}
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
