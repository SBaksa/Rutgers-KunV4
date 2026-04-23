package verification

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/email"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/validation"
	"github.com/bwmarrin/discordgo"
)

// HandleDMMessage processes DM messages for verification flow
func HandleDMMessage(s *discordgo.Session, m *discordgo.MessageCreate, log *logger.Logger) {
	// Only handle DMs
	if m.GuildID != "" {
		return
	}

	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Get verification state
	var verificationState VerificationState
	err := database.Instance.GetAgreementState(m.Author.ID, &verificationState)
	if err != nil {
		log.Debug("No verification state found", "user", m.Author.ID)
		return
	}

	log.Debug("Processing verification step", "user", m.Author.ID, "step", verificationState.Step)

	// Check if expired
	if verificationState.IsExpired() {
		cleanupErr := database.Instance.RemoveAgreementState(m.Author.ID)
		if cleanupErr != nil {
			log.Error("Failed to remove expired agreement state", "user", m.Author.ID, "error", cleanupErr)
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Session Expired",
			Color:       0xFF0000,
			Description: "Your verification session has expired. Please start over with `!agree` in your server.",
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	switch verificationState.Step {
	case StepRoleSelection:
		handleRoleSelection(s, m, &verificationState, log)
	case StepNetIDEntry:
		handleNetIDEntry(s, m, &verificationState, log)
	case StepCodeEntry:
		handleCodeEntry(s, m, &verificationState, log)
	default:
		log.Warn("Unknown verification step", "user", m.Author.ID, "step", verificationState.Step)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please start over with `!agree`.")
	}
}

// handleRoleSelection processes role selection step
func handleRoleSelection(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState, log *logger.Logger) {
	roleName := strings.ToLower(strings.TrimSpace(m.Content))

	// Get agreement roles for the guild
	agreementRoles, err := database.Instance.GetAgreementRoles(state.GuildID)
	if err != nil {
		log.Error("Failed to get agreement roles", "guild", state.GuildID, "error", err)
		s.ChannelMessageSend(m.ChannelID, "Error retrieving roles. Please try again.")
		return
	}

	if len(agreementRoles) == 0 {
		log.Warn("No agreement roles configured", "guild", state.GuildID)
		s.ChannelMessageSend(m.ChannelID, "No roles are available in this server.")
		return
	}

	// Find matching role
	var selectedRole *database.AgreementRole
	for i, role := range agreementRoles {
		if role.Authenticate == "permission" {
			continue // Skip permission roles
		}

		// Get Discord role to compare names
		discordRole, discordErr := s.State.Role(state.GuildID, role.RoleID)
		if discordErr != nil {
			log.Debug("Failed to get Discord role info", "roleID", role.RoleID, "error", discordErr)
			continue
		}

		if strings.ToLower(discordRole.Name) == roleName {
			selectedRole = &agreementRoles[i]
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
			if discordRole, discordErr := s.State.Role(state.GuildID, role.RoleID); discordErr == nil {
				validRoles = append(validRoles, discordRole.Name)
			}
		}

		log.Info("Invalid role selected", "user", m.Author.ID, "requested", roleName)
		embed := &discordgo.MessageEmbed{
			Title:       "Invalid Role",
			Color:       0xFF0000,
			Description: fmt.Sprintf("Please choose from the available roles:\n```\n%s\n```", strings.Join(validRoles, "\n")),
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	log.Info("Role selected", "user", m.Author.ID, "role", selectedRole.RoleID)

	// Check if role requires verification
	if selectedRole.Authenticate == "false" {
		// No verification needed, assign role immediately
		assignRoleAndFinish(s, m, state, selectedRole.RoleID, log)
		return
	}

	// Role requires verification, move to NetID step
	state.SetRole(selectedRole.RoleID)
	if stateErr := database.Instance.SetAgreementState(m.Author.ID, state); stateErr != nil {
		log.Error("Failed to update agreement state", "user", m.Author.ID, "error", stateErr)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try `!agree` again.")
		return
	}

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
func handleNetIDEntry(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState, log *logger.Logger) {
	netID := validation.NormalizeNetID(m.Content)

	// Validate NetID format
	if !validation.IsValidNetID(netID) {
		log.Debug("Invalid NetID format", "user", m.Author.ID, "netID", netID)
		s.ChannelMessageSend(m.ChannelID, "That doesn't appear to be a valid NetID. Please re-enter your NetID (e.g., `sab468` or `sb468`).")
		return
	}

	// Generate verification code
	verificationCode := email.GenerateVerificationCode()

	// Get guild info for email
	guild, guildErr := s.Guild(state.GuildID)
	if guildErr != nil {
		log.Error("Failed to get guild info", "guild", state.GuildID, "error", guildErr)
		s.ChannelMessageSend(m.ChannelID, "Error retrieving server information. Please try again.")
		return
	}

	// Update state before sending email
	state.SetNetID(netID, verificationCode)
	if stateErr := database.Instance.SetAgreementState(m.Author.ID, state); stateErr != nil {
		log.Error("Failed to update agreement state", "user", m.Author.ID, "error", stateErr)
		s.ChannelMessageSend(m.ChannelID, "An error occurred. Please try `!agree` again.")
		return
	}

	// Send verification email
	smtpConfig := email.LoadSMTPConfig()
	if !smtpConfig.IsConfigured() {
		log.Error("SMTP not configured", "user", m.Author.ID)
		s.ChannelMessageSend(m.ChannelID, "Email verification is not configured. Please contact an administrator.")
		return
	}

	// Send email
	emailErr := smtpConfig.SendVerificationEmail(netID, verificationCode, guild.Name)
	if emailErr != nil {
		log.Error("Failed to send verification email", "user", m.Author.ID, "netID", netID, "error", emailErr)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to send verification email.\n\nPlease check that you entered your NetID correctly or contact an administrator."))
		// Clean up state on email failure
		_ = database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	log.Info("Verification email sent", "user", m.Author.ID, "netID", netID)

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
func handleCodeEntry(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState, log *logger.Logger) {
	enteredCode := strings.TrimSpace(m.Content)

	// Verify code matches
	if enteredCode != state.Code {
		log.Debug("Invalid verification code", "user", m.Author.ID)
		s.ChannelMessageSend(m.ChannelID, "That doesn't appear to be the right verification code. Make sure you're entering it correctly.")
		return
	}

	log.Info("Verification code valid", "user", m.Author.ID)

	// Code is correct, assign role
	assignRoleAndFinish(s, m, state, state.RoleID, log)
}

// assignRoleAndFinish assigns the role and completes verification
func assignRoleAndFinish(s *discordgo.Session, m *discordgo.MessageCreate, state *VerificationState, roleID string, log *logger.Logger) {
	// Get guild member
	_, memberErr := s.GuildMember(state.GuildID, m.Author.ID)
	if memberErr != nil {
		log.Warn("User not found in guild", "user", m.Author.ID, "guild", state.GuildID, "error", memberErr)
		s.ChannelMessageSend(m.ChannelID, "You could not be found in the server. Please make sure you're still in the server and try `!agree` again.")
		_ = database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Get role info
	role, roleErr := s.State.Role(state.GuildID, roleID)
	if roleErr != nil {
		log.Error("Failed to get role info", "roleID", roleID, "guild", state.GuildID, "error", roleErr)
		s.ChannelMessageSend(m.ChannelID, "Error retrieving role information. Please contact an administrator.")
		_ = database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	// Assign the role
	if assignErr := s.GuildMemberRoleAdd(state.GuildID, m.Author.ID, roleID); assignErr != nil {
		log.Error("Failed to assign role", "user", m.Author.ID, "role", roleID, "guild", state.GuildID, "error", assignErr)
		s.ChannelMessageSend(m.ChannelID, "Failed to assign role. Please contact an administrator.")
		_ = database.Instance.RemoveAgreementState(m.Author.ID)
		return
	}

	log.Info("Role assigned", "user", m.Author.ID, "role", role.Name, "guild", state.GuildID)

	// Get permission role if it exists and assign it too
	agreementRoles, rolesErr := database.Instance.GetAgreementRoles(state.GuildID)
	if rolesErr == nil {
		for _, ar := range agreementRoles {
			if ar.Authenticate == "permission" {
				if permErr := s.GuildMemberRoleAdd(state.GuildID, m.Author.ID, ar.RoleID); permErr != nil {
					log.Warn("Failed to assign permission role", "roleID", ar.RoleID, "error", permErr)
				} else {
					log.Debug("Permission role assigned", "roleID", ar.RoleID, "user", m.Author.ID)
				}
				break
			}
		}
	}

	// Clean up verification state
	if cleanupErr := database.Instance.RemoveAgreementState(m.Author.ID); cleanupErr != nil {
		log.Error("Failed to clean up agreement state", "user", m.Author.ID, "error", cleanupErr)
	}

	// Success message
	guild, _ := s.Guild(state.GuildID)
	guildName := "the server"
	if guild != nil {
		guildName = guild.Name
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Verification Complete!",
		Color:       0x00FF00,
		Description: fmt.Sprintf("You have successfully been given the **%s** role in **%s**!", role.Name, guildName),
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)

	log.Info("Verification completed successfully", "user", m.Author.ID, "guild", state.GuildID)

	welcomeChannelID, err := database.Instance.GetGuildSettingString(state.GuildID, "welcomeChannel")
	if err == nil && welcomeChannelID != "" {
		welcomeText, textErr := database.Instance.GetGuildSettingString(state.GuildID, "welcomeText")
		if textErr != nil || welcomeText == "" {
			welcomeText = fmt.Sprintf("Welcome to %s, <@%s>!", guildName, m.Author.ID)
		} else {
			welcomeText = strings.ReplaceAll(welcomeText, "[user]", fmt.Sprintf("<@%s>", m.Author.ID))
			welcomeText = strings.ReplaceAll(welcomeText, "[guild]", guildName)
		}
		s.ChannelMessageSend(welcomeChannelID, welcomeText)
	}

}
