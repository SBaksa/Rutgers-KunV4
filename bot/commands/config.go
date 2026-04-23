package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func SetWelcomeChannel(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!setwelcomechannel #channel` or `!setwelcomechannel clear`")
		return err
	}

	if args[0] == "clear" {
		database.Instance.RemoveGuildSetting(m.GuildID, "welcomeChannel")
		_, err := s.ChannelMessageSend(m.ChannelID, "Welcome channel successfully removed.")
		return err
	}

	channelID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(args[0], "<#"), ">"), "!")
	database.Instance.SetGuildSetting(m.GuildID, "welcomeChannel", channelID)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Welcome channel successfully set as <#%s>.", channelID))
	return err
}

func SetWelcomeText(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!setwelcometext <text>` — use `[user]` and `[guild]` as placeholders. Or `!setwelcometext clear`")
		return err
	}

	if args[0] == "clear" {
		database.Instance.RemoveGuildSetting(m.GuildID, "welcomeText")
		_, err := s.ChannelMessageSend(m.ChannelID, "Welcome text successfully removed.")
		return err
	}

	agreementChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "agreementChannel")
	if err != nil || agreementChannel == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "You need to set an agreement channel first with `!setagreementchannel`.")
		return err
	}

	welcomeText := strings.Join(args, " ")
	database.Instance.SetGuildSetting(m.GuildID, "welcomeText", welcomeText)
	_, err = s.ChannelMessageSend(m.ChannelID, "Welcome text successfully set.")
	return err
}

func SetLogChannel(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!setlogchannel #channel` or `!setlogchannel clear`")
		return err
	}

	if args[0] == "clear" {
		database.Instance.RemoveGuildSetting(m.GuildID, "logChannel")
		_, err := s.ChannelMessageSend(m.ChannelID, "Log channel successfully removed.")
		return err
	}

	channelID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(args[0], "<#"), ">"), "!")
	database.Instance.SetGuildSetting(m.GuildID, "logChannel", channelID)
	_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Log channel successfully set as <#%s>.", channelID))
	return err
}

func SetAgreementChannel(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!setagreementchannel #channel` or `!setagreementchannel clear`")
		return err
	}

	if args[0] == "clear" {
		database.Instance.RemoveGuildSetting(m.GuildID, "agreementChannel")
		_, err := s.ChannelMessageSend(m.ChannelID, "Agreement channel successfully removed.")
		return err
	}

	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err != nil || len(agreementRoles) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "You need to set up agreement roles first with `!setagreementroles`.")
		return err
	}

	welcomeChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "welcomeChannel")
	if err != nil || welcomeChannel == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, "You need to set up the welcome channel first with `!setwelcomechannel`.")
		return err
	}

	channelID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(args[0], "<#"), ">"), "!")
	database.Instance.SetAgreementChannel(m.GuildID, channelID)
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Agreement channel successfully set as <#%s>.", channelID))
	return err
}

func SetAgreementRoles(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}
	if len(args) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Usage: `!setagreementroles <role name>, <true/false/permission>`\nExample: `!setagreementroles Student, true`\nUse `clear` to reset.")
		return err
	}

	if args[0] == "clear" {
		database.Instance.RemoveGuildSetting(m.GuildID, "agreementRoles")
		_, err := s.ChannelMessageSend(m.ChannelID, "Agreement roles successfully cleared.")
		return err
	}

	input := strings.Join(args, " ")
	parts := strings.SplitN(input, ",", 2)
	if len(parts) != 2 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Invalid format. Use `!setagreementroles <role name>, <true/false/permission>`")
		return err
	}

	roleName := strings.TrimSpace(strings.ToLower(parts[0]))
	authenticate := strings.TrimSpace(strings.ToLower(parts[1]))

	if authenticate != "true" && authenticate != "false" && authenticate != "permission" {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Authentication must be `true`, `false`, or `permission`.")
		return err
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Failed to fetch server info.")
		return sendErr
	}

	var foundRoleID string
	for _, role := range guild.Roles {
		if strings.ToLower(role.Name) == roleName {
			foundRoleID = role.ID
			break
		}
	}

	if foundRoleID == "" {
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Role `%s` not found.", roleName))
		return err
	}

	existingRoles, _ := database.Instance.GetAgreementRoles(m.GuildID)

	permissionCount := 0
	for _, r := range existingRoles {
		if r.Authenticate == "permission" {
			permissionCount++
		}
	}
	if authenticate == "permission" && permissionCount >= 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ Only one permission role is allowed.")
		return err
	}

	existingRoles = append(existingRoles, database.AgreementRole{
		RoleID:       foundRoleID,
		Authenticate: authenticate,
	})

	database.Instance.SetAgreementRoles(m.GuildID, existingRoles)
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Role `%s` added to agreement roles with authentication: `%s`.", roleName, authenticate))
	return err
}

func ListConfig(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	if m.GuildID == "" {
		return nil
	}
	if !isModerator(s, m) {
		_, err := s.ChannelMessageSend(m.ChannelID, "❌ You don't have permission to use this command.")
		return err
	}

	guild, err := s.Guild(m.GuildID)
	if err != nil {
		_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ Failed to fetch server info.")
		return sendErr
	}

	embed := &discordgo.MessageEmbed{
		Title:     "Configs for " + guild.Name,
		Color:     0xCC0033,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: guild.IconURL("")},
	}

	if agreementChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "agreementChannel"); err == nil && agreementChannel != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Agreement Channel", Value: fmt.Sprintf("<#%s>", agreementChannel)})
	}
	if welcomeChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "welcomeChannel"); err == nil && welcomeChannel != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Welcome Channel", Value: fmt.Sprintf("<#%s>", welcomeChannel)})
	}
	if welcomeText, err := database.Instance.GetGuildSettingString(m.GuildID, "welcomeText"); err == nil && welcomeText != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Welcome Text", Value: welcomeText})
	}
	if logChannel, err := database.Instance.GetGuildSettingString(m.GuildID, "logChannel"); err == nil && logChannel != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Log Channel", Value: fmt.Sprintf("<#%s>", logChannel)})
	}

	agreementRoles, err := database.Instance.GetAgreementRoles(m.GuildID)
	if err == nil && len(agreementRoles) > 0 {
		var roleList []string
		for _, role := range agreementRoles {
			roleList = append(roleList, fmt.Sprintf("<@&%s>, %s", role.RoleID, role.Authenticate))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Agreement Roles", Value: strings.Join(roleList, "\n")})
	}

	if len(embed.Fields) == 0 {
		_, err := s.ChannelMessageSend(m.ChannelID, "No configs set for this server. Use `!help` to get started.")
		return err
	}

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
