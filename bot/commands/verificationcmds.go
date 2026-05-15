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
	if !IsModerator(s, m) {
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

	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!roleswitch <role name or @role>`")
		return err
	}

	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err != nil || len(agreementRoles) == 0 {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "No agreement roles configured for this server.")
		return sendErr
	}

	var verifiableRoles []database.AgreementRole
	for _, ar := range agreementRoles {
		if ar.Authenticate != "permission" {
			verifiableRoles = append(verifiableRoles, ar)
		}
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

	var prevRole *database.AgreementRole
	for i, ar := range verifiableRoles {
		if memberRoleSet[ar.RoleID] {
			prevRole = &verifiableRoles[i]
			break
		}
	}

	if prevRole == nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "You don't have any agreement roles to switch from. Use `!agree` to get started.")
		return sendErr
	}

	// Resolve target role — try mention first, then name
	var toRole *database.AgreementRole
	targetArg := strings.ToLower(strings.Join(args, " "))
	if strings.HasPrefix(args[0], "<@&") && strings.HasSuffix(args[0], ">") {
		mentionID := strings.TrimSuffix(strings.TrimPrefix(args[0], "<@&"), ">")
		for i, ar := range verifiableRoles {
			if ar.RoleID == mentionID {
				toRole = &verifiableRoles[i]
				break
			}
		}
	} else {
		for i, ar := range verifiableRoles {
			discordRole, roleErr := s.State.Role(m.GuildID, ar.RoleID)
			if roleErr == nil && strings.ToLower(discordRole.Name) == targetArg {
				toRole = &verifiableRoles[i]
				break
			}
		}
	}

	if toRole == nil {
		var roleNames []string
		for _, ar := range verifiableRoles {
			if ar.RoleID == prevRole.RoleID {
				continue
			}
			if role, roleErr := s.State.Role(m.GuildID, ar.RoleID); roleErr == nil {
				roleNames = append(roleNames, role.Name)
			}
		}
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Make sure you enter one of these roles: %s.", strings.Join(roleNames, ", ")))
		return sendErr
	}

	if toRole.RoleID == prevRole.RoleID || memberRoleSet[toRole.RoleID] {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "You already have this role.")
		return sendErr
	}

	prevName := prevRole.RoleID
	toName := toRole.RoleID
	if r, e := s.State.Role(m.GuildID, prevRole.RoleID); e == nil {
		prevName = r.Name
	}
	if r, e := s.State.Role(m.GuildID, toRole.RoleID); e == nil {
		toName = r.Name
	}

	// Unverified → verified: need email verification via DM
	if prevRole.Authenticate == "false" && toRole.Authenticate == "true" {
		state, stateErr := vm.StartVerification(m.Author.ID, m.GuildID)
		if stateErr != nil {
			_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not start role switch. Please try again.")
			return sendErr
		}
		state.SetRole(toRole.RoleID)
		state.RemoveRole = prevRole.RoleID
		state.NoWelcome = true
		if updateErr := vm.UpdateVerificationState(m.Author.ID, state); updateErr != nil {
			_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not start role switch. Please try again.")
			return sendErr
		}

		dmChannel, dmErr := s.UserChannelCreate(m.Author.ID)
		if dmErr != nil {
			_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Could not open a DM with you. Make sure DMs are enabled.")
			return sendErr
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Email Verification Required",
			Color:       0xCC0033,
			Description: fmt.Sprintf("In order to switch to **%s** you need to complete email verification.\n\nPlease enter your Rutgers NetID.", toName),
		}
		if _, dmSendErr := s.ChannelMessageSendEmbed(dmChannel.ID, embed); dmSendErr != nil {
			return dmSendErr
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "You've been DM'd instructions on switching your role.")
		return err
	}

	// All other cases: direct swap
	if removeErr := s.GuildMemberRoleRemove(m.GuildID, m.Author.ID, prevRole.RoleID); removeErr != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error removing role %s: %v. Make sure the bot has Manage Roles permission.", prevName, removeErr))
		return sendErr
	}
	if addErr := s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, toRole.RoleID); addErr != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error adding role %s: %v. Make sure the bot has Manage Roles permission.", toName, addErr))
		return sendErr
	}

	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Successfully switched your role to **%s**.", toName))
	return err
}
