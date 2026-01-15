package verification

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/email"
	"github.com/SBaksa/Rutgers-KunV4/validation"
	"github.com/bwmarrin/discordgo"
)

// HandleDMMessage processes DM messages for verification flow
func HandleDMMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Only handle DMs
	if m.GuildID != "" {
		return
	}

	// Get verification state
	var verificationState VerificationState
	err := database.Instance.GetAgreementState(m.Author.ID, &verificationState)
	if err != nil {
		// No verification in progress
		return
	}

	// Check if expired
	if verificationState.IsExpired() {
		database.Instance.RemoveAgreementState(m.Author.ID)
		s.ChannelMessageSend(m.ChannelID, "Your verification session has expired. Please start over with `!agree`.")
		return
	}

	switch verificationState.Step {
	case StepRoleSelection:
		handleRoleSelection(s, m, &verificationState)
	case StepNetIDEntry:
		handleNetIDEntry(s, m, &verificationState)
	case StepCodeEntry:
		handleCodeEntry(s, m, &verificationState)
	}
}

// handleRoleSelection processes role selection step
func handleRoleSelection(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState) {
	roleName := strings.ToLower(strings.TrimSpace(m.Content))

	// Get agreement roles for the guild
	agreementRoles, err := database.Instance.GetAgreementRoles(state.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving roles. Please try again.")
		return
	}

	// Find matching role
	var selectedRole *database.AgreementRole
	for _, role := range agreementRoles {
		if role.Authenticate == "permission" {
			continue // Skip permission roles
		}

		// Get Discord role to compare names
		discordRole, err := s.State.Role(state.GuildID, role.RoleID)
		if err != nil {
			continue
		}

		if strings.ToLower(discordRole.Name) == roleName {
			selectedRole = &role
			break
		}
	}

	if selectedRole == nil {
		// Get valid role names for error message
		var validRoles []string
		for _, role := range agreementRoles {
			if role.Authenticate == "permission" {
				continue
			}
			if discordRole, err := s.State.Role(state.GuildID, role.RoleID); err == nil {
				validRoles = append(validRoles, discordRole.Name)
			}
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid role. Please choose from: %s", strings.Join(validRoles, ", ")))
		return
	}

	// Check if role requires verification
	if selectedRole.Authenticate == "false" {
		// No verification needed, assign role immediately
		assignRoleAndFinish(s, m, state, selectedRole.RoleID)
		return
	}

	// Role requires verification, move to NetID step
	state.SetRole(selectedRole.RoleID)
	database.Instance.SetAgreementState(m.Author.ID, state)

	embed := &discordgo.MessageEmbed{
		Title:       "NetID Entry",
		Color:       0xCC0033,
		Description: "Now enter your Rutgers NetID. Your NetID is a unique identifier given to you by Rutgers that you use to sign in to all your Rutgers services.\n\nIt is generally your initials followed by a few numbers (e.g., `sab468` or `sb468`).",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "This verification will expire in 15 minutes",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// handleNetIDEntry processes NetID entry step
func handleNetIDEntry(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState) {
	netID := validation.NormalizeNetID(m.Content)

	// Validate NetID format
	if !validation.IsValidNetID(netID) {
		s.ChannelMessageSend(m.ChannelID, "That doesn't appear to be a valid NetID. Please re-enter your NetID (e.g., `sab468` or `sb468`).")
		return
	}

	// Generate verification code
	verificationCode := email.GenerateVerificationCode()

	// Get guild info for email
	guild, err := s.Guild(state.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving server information. Please try again.")
		return
	}

	// Send verification email
	smtpConfig := email.LoadSMTPConfig()

	// Update state first
	state.SetNetID(netID, verificationCode)
	database.Instance.SetAgreementState(m.Author.ID, state)

	// Send email
	err = smtpConfig.SendVerificationEmail(netID, verificationCode, guild.Name)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to send verification email: %v\n\nPlease check that you entered your NetID correctly.", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Verification Code Sent",
		Color:       0x00FF00,
		Description: fmt.Sprintf("Email successfully sent to `%s@scarletmail.rutgers.edu`.\n\nPlease check your school email for a verification code and enter it here.\n\nIf you don't receive an email, you may have entered your NetID incorrectly.", netID),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Code expires in 15 minutes",
		},
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// handleCodeEntry processes verification code entry step
func handleCodeEntry(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState) {
	enteredCode := strings.TrimSpace(m.Content)

	// Verify code matches
	if enteredCode != state.Code {
		s.ChannelMessageSend(m.ChannelID, "That doesn't appear to be the right verification code. Make sure you're entering it correctly.")
		return
	}

	// Code is correct, assign role
	assignRoleAndFinish(s, m, state, state.RoleID)
}

// assignRoleAndFinish assigns the role and completes verification
func assignRoleAndFinish(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState, roleID string) {
	// Get guild member
	_, err := s.GuildMember(state.GuildID, m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "You could not be found in the server. Please make sure you're still in the server and try `!agree` again.")
		database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Get role info
	role, err := s.State.Role(state.GuildID, roleID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error retrieving role information. Please contact an administrator.")
		database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Assign the role
	err = s.GuildMemberRoleAdd(state.GuildID, m.Author.ID, roleID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to assign role. Please contact an administrator.")
		database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Get permission role if it exists and assign it too
	agreementRoles, err := database.Instance.GetAgreementRoles(state.GuildID)
	if err == nil {
		for _, ar := range agreementRoles {
			if ar.Authenticate == "permission" {
				s.GuildMemberRoleAdd(state.GuildID, m.Author.ID, ar.RoleID)
				break
			}
		}
	}

	// Clean up verification state
	database.Instance.RemoveAgreementState(m.Author.ID)

	// Success message
	guild, _ := s.Guild(state.GuildID)
	guildName := "the server"
	if guild != nil {
		guildName = guild.Name
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Verification Complete!",
		Color:       0x00FF00,
		Description: fmt.Sprintf("You have successfully been given the **%s** role in **%s**!", role.Name, guildName),
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)

	// TODO: Send welcome message to welcome channel if configured
}
