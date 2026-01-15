package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/email"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Agree(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Only works in guilds
	if m.GuildID == "" {
		return
	}

	// Check if agreement channel is set
	agreementChannelID, err := database.Instance.GetAgreementChannel(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Agreement channel not configured. Please ask an admin to set it up.")
		return
	}

	// Check if this is the agreement channel
	if m.ChannelID != agreementChannelID {
		s.ChannelMessageSend(m.ChannelID, "This command can only be used in the agreement channel.")
		return
	}

	// Check if agreement roles are configured
	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err != nil || len(agreementRoles) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Agreement roles not configured. Please ask an admin to set them up.")
		return
	}

	// Check if SMTP is configured
	smtpConfig := email.LoadSMTPConfig()
	if !smtpConfig.IsConfigured() {
		s.ChannelMessageSend(m.ChannelID, "Email verification is not configured. Please contact the bot administrator.")
		return
	}

	// Filter out permission roles and get user-selectable roles
	selectableRoles := make([]database.AgreementRole, 0)
	for _, role := range agreementRoles {
		if role.Authenticate != "permission" {
			selectableRoles = append(selectableRoles, role)
		}
	}

	if len(selectableRoles) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No user roles are configured for agreement.")
		return
	}

	// Get role names for display
	var roleNames []string
	for _, role := range selectableRoles {
		// Get the actual role from Discord to get its name
		discordRole, err := s.State.Role(m.GuildID, role.RoleID)
		if err != nil {
			continue // Skip roles that don't exist
		}
		roleNames = append(roleNames, discordRole.Name)
	}

	if len(roleNames) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No valid roles found. Please ask an admin to check the role configuration.")
		return
	}

	// Start verification process
	verificationState := verification.NewVerificationState(m.GuildID)

	// Store verification state in database
	err = database.Instance.SetAgreementState(m.Author.ID, verificationState)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to start verification process. Please try again.")
		return
	}

	// Send DM to user
	dmChannel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "I couldn't send you a DM. Please make sure you allow DMs from server members.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Role Selection",
		Color: 0xCC0033,
		Description: fmt.Sprintf("Please enter the name of the role you want to add. Available roles are:\n\n%s",
			strings.Join(roleNames, "\n")),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "This verification will expire in 15 minutes",
		},
	}

	_, err = s.ChannelMessageSendEmbed(dmChannel.ID, embed)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "I couldn't send you a DM. Please make sure you allow DMs from server members.")
		// Clean up verification state
		database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Success message in channel
	s.ChannelMessageSend(m.ChannelID, "Check your DMs to continue the verification process!")
}
