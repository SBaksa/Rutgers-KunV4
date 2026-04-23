package commands

import (
	"fmt"
	"strings"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func WhoIs(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	var targetUser *discordgo.User

	if len(args) == 0 {
		targetUser = m.Author
	} else {
		userMention := args[0]
		userID := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(userMention, "<@"), ">"), "!")

		var err error
		targetUser, err = s.User(userID)
		if err != nil {
			_, sendErr := s.ChannelMessageSend(m.ChannelID, "❌ User not found. Try mentioning them with @user")
			return sendErr
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: targetUser.Username + "#" + targetUser.Discriminator,
		Color: 0x0099FF,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: targetUser.AvatarURL("512"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "User ID",
				Value:  targetUser.ID,
				Inline: true,
			},
		},
	}

	timestamp, _ := discordgo.SnowflakeTimestamp(targetUser.ID)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Account Created",
		Value:  timestamp.Format("January 2, 2006"),
		Inline: true,
	})

	if targetUser.Bot {
		embed.Description = "🤖 This is a bot!"
	}

	if targetUser.ID == s.State.User.ID {
		prefix := "!"
		embed.Description = fmt.Sprintf("If you want to find out more about me, use `%swhoami`", prefix)
	}

	if m.GuildID != "" {
		member, err := s.GuildMember(m.GuildID, targetUser.ID)
		if err == nil {
			if len(member.Roles) > 0 {
				guild, guildErr := s.Guild(m.GuildID)
				if guildErr == nil {
					var roleNames []string
					for _, roleID := range member.Roles {
						for _, role := range guild.Roles {
							if role.ID == roleID && role.ID != m.GuildID {
								roleNames = append(roleNames, role.Name)
							}
						}
					}
					if len(roleNames) > 0 {
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:  "Roles",
							Value: strings.Join(roleNames, "\n"),
						})
					}
				}
			}

			if !member.JoinedAt.IsZero() {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "Joined Server",
					Value:  member.JoinedAt.Format("January 2, 2006"),
					Inline: true,
				})
			}
		}
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}
