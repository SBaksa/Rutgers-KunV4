package commands

import (
	"fmt"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/SBaksa/Rutgers-KunV4/verification"
	"github.com/bwmarrin/discordgo"
)

func Ping(s *discordgo.Session, m *discordgo.MessageCreate, args []string, log *logger.Logger, vm *verification.VerificationManager) error {
	before := time.Now()
	msg, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		return err
	}
	roundTrip := time.Since(before).Milliseconds()
	heartbeat := s.HeartbeatLatency().Milliseconds()

	_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, fmt.Sprintf("Pong! Round-trip: **%dms** | Heartbeat: **%dms**", roundTrip, heartbeat))
	return err
}
