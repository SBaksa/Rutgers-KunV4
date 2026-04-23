package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Diagnose(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}

	embed := &discordgo.MessageEmbed{
		Title: "Verification Diagnostics",
		Color: 0xCC0033,
	}

	agreementChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "agreementChannel")
	if err != nil || agreementChannel == "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "❌ Agreement Channel",
			Value: "Not set. Use `!setagreementchannel #channel`",
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "✅ Agreement Channel",
			Value: fmt.Sprintf("<#%s>", agreementChannel),
		})
	}

	agreementRoles, rolesErr := database.Instance.GetAgreementRoles(m.GuildID)
	if rolesErr != nil || len(agreementRoles) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "❌ Agreement Roles",
			Value: "Not set. Use `!setagreementroles`",
		})
	} else {
		var roleList []string
		for _, role := range agreementRoles {
			roleList = append(roleList, fmt.Sprintf("<@&%s> (%s)", role.RoleID, role.Authenticate))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "✅ Agreement Roles",
			Value: strings.Join(roleList, "\n"),
		})
	}

	welcomeChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "welcomeChannel")
	if err != nil || welcomeChannel == "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "⚠️ Welcome Channel",
			Value: "Not set. Use `!setwelcomechannel #channel`",
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "✅ Welcome Channel",
			Value: fmt.Sprintf("<#%s>", welcomeChannel),
		})
	}

	stats := vm.GetStats()
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:  "Active Verifications",
		Value: fmt.Sprintf("%v", stats["active_verifications"]),
	})

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func RoleSwitch(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}

	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err != nil || len(agreementRoles) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "No agreement roles configured for this server.")
		return sendErr
	}

	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not fetch your member info.")
		return sendErr
	}

	memberRoleSet := make(map[string]bool)
	for _, roleID := range member.Roles {
		memberRoleSet[roleID] = true
	}

	var currentRole *database.AgreementRole
	for i, ar := range agreementRoles {
		if ar.Authenticate == "permission" {
			continue
		}
		if memberRoleSet[ar.RoleID] {
			currentRole = &agreementRoles[i]
			break
		}
	}

	if currentRole == nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "You don't currently have any agreement roles to switch from. Use `!agree` to get started.")
		return sendErr
	}

	state, stateErr := vm.StartVerification(m.Author.ID, m.GuildID)
	if stateErr != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not start role switch. Please try again.")
		return sendErr
	}
	state.RemoveRole = currentRole.RoleID

	if updateErr := vm.UpdateVerificationState(m.Author.ID, state); updateErr != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not start role switch. Please try again.")
		return sendErr
	}

	dmChannel, dmErr := s.UserChannelCreate(m.Author.ID)
	if dmErr != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not open a DM with you. Make sure DMs are enabled.")
		return sendErr
	}

	var roleList []string
	for _, ar := range agreementRoles {
		if ar.Authenticate == "permission" || ar.RoleID == currentRole.RoleID {
			continue
		}
		role, roleErr := s.State.Role(m.GuildID, ar.RoleID)
		if roleErr == nil {
			roleList = append(roleList, role.Name)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Role Switch",
		Color:       0xCC0033,
		Description: fmt.Sprintf("Which role would you like to switch to?\n\nAvailable roles:\n%s", strings.Join(roleList, "\n")),
	}

	_, err = s.ChannelMessageSendEmbed(dmChannel.ID, embed)
	return err
}
