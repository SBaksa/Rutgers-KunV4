package verification

import (
	"sync"
	"time"

	"github.com/SBaksa/Rutgers-KunV4/database"
	"github.com/SBaksa/Rutgers-KunV4/logger"
	"github.com/bwmarrin/discordgo"
)

// VerificationManager coordinates the verification process
type VerificationManager struct {
	log                 *logger.Logger
	activeVerifications map[string]*VerificationState
	mu                  sync.RWMutex
}

// NewVerificationManager creates a new verification manager
func NewVerificationManager(log *logger.Logger) *VerificationManager {
	return &VerificationManager{
		log:                 log,
		activeVerifications: make(map[string]*VerificationState),
	}
}

// StartVerification initiates a verification flow for a user
func (vm *VerificationManager) StartVerification(userID, guildID string) (*VerificationState, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	state := NewVerificationState(guildID)

	if err := database.Instance.SetAgreementState(userID, state); err != nil {
		vm.log.Error("Failed to save verification state", "user", userID, "error", err)
		return nil, err
	}

	vm.activeVerifications[userID] = state
	vm.log.Info("Verification started", "user", userID, "guild", guildID)

	return state, nil
}

// GetVerificationState retrieves active verification state
func (vm *VerificationManager) GetVerificationState(userID string) (*VerificationState, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	state, exists := vm.activeVerifications[userID]
	return state, exists
}

// CompleteVerification marks verification as complete and cleans up
func (vm *VerificationManager) CompleteVerification(userID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	delete(vm.activeVerifications, userID)

	if err := database.Instance.RemoveAgreementState(userID); err != nil {
		vm.log.Error("Failed to remove verification state", "user", userID, "error", err)
		return err
	}

	vm.log.Info("Verification completed", "user", userID)
	return nil
}

// CancelVerification cancels an ongoing verification
func (vm *VerificationManager) CancelVerification(userID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	delete(vm.activeVerifications, userID)

	if err := database.Instance.RemoveAgreementState(userID); err != nil {
		vm.log.Error("Failed to cancel verification state", "user", userID, "error", err)
		return err
	}

	vm.log.Info("Verification cancelled", "user", userID)
	return nil
}

// UpdateVerificationState updates an ongoing verification state
func (vm *VerificationManager) UpdateVerificationState(userID string, state *VerificationState) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if err := database.Instance.SetAgreementState(userID, state); err != nil {
		vm.log.Error("Failed to update verification state", "user", userID, "error", err)
		return err
	}

	vm.activeVerifications[userID] = state
	return nil
}

// ProcessDMMessage handles incoming DM for verification
func (vm *VerificationManager) ProcessDMMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Only handle DMs
	if m.GuildID != "" {
		return
	}

	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	HandleDMMessage(s, m, vm.log, vm)
}

// CleanupExpiredVerifications removes expired verification states
// Should be called periodically (e.g., every minute)
func (vm *VerificationManager) CleanupExpiredVerifications() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	now := time.Now()
	expiredUsers := []string{}

	for userID, state := range vm.activeVerifications {
		if now.After(state.ExpiresAt) {
			expiredUsers = append(expiredUsers, userID)
		}
	}

	for _, userID := range expiredUsers {
		delete(vm.activeVerifications, userID)
		if err := database.Instance.RemoveAgreementState(userID); err != nil {
			vm.log.Error("Failed to remove expired verification", "user", userID, "error", err)
		} else {
			vm.log.Debug("Expired verification cleaned up", "user", userID)
		}
	}
}

// GetStats returns verification manager statistics
func (vm *VerificationManager) GetStats() map[string]interface{} {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	return map[string]interface{}{
		"active_verifications": len(vm.activeVerifications),
	}
}
