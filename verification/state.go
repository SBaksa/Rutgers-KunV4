package verification

import (
	"time"
)

// VerificationStep represents where the user is in the verification process
type VerificationStep int

const (
	StepRoleSelection VerificationStep = 1
	StepNetIDEntry    VerificationStep = 2
	StepCodeEntry     VerificationStep = 3
)

// VerificationState tracks a user's verification progress
type VerificationState struct {
	GuildID    string           `json:"guildID"`
	Step       VerificationStep `json:"step"`
	RoleID     string           `json:"roleID,omitempty"`
	NetID      string           `json:"netID,omitempty"`
	Code       string           `json:"code,omitempty"`
	ExpiresAt  time.Time        `json:"expiresAt"`
	NoWelcome  bool             `json:"nowelcome,omitempty"`
	RemoveRole string           `json:"removerole,omitempty"`
}

// IsExpired checks if the verification state has expired
func (vs *VerificationState) IsExpired() bool {
	return time.Now().After(vs.ExpiresAt)
}

// NewVerificationState creates a new verification state
func NewVerificationState(guildID string) *VerificationState {
	return &VerificationState{
		GuildID:   guildID,
		Step:      StepRoleSelection,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15 minute timeout
	}
}

// SetRole moves to NetID entry step
func (vs *VerificationState) SetRole(roleID string) {
	vs.RoleID = roleID
	vs.Step = StepNetIDEntry
	vs.ExpiresAt = time.Now().Add(15 * time.Minute) // Reset timeout
}

// SetNetID moves to code entry step
func (vs *VerificationState) SetNetID(netID, code string) {
	vs.NetID = netID
	vs.Code = code
	vs.Step = StepCodeEntry
	vs.ExpiresAt = time.Now().Add(15 * time.Minute) // Reset timeout
}

// IsComplete checks if verification is ready for role assignment
func (vs *VerificationState) IsComplete() bool {
	return vs.Step == StepCodeEntry && vs.RoleID != "" && vs.NetID != "" && vs.Code != ""
}
