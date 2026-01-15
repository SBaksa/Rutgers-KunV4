package database

// SettingsProvider defines the interface for bot settings storage
// This allows us to swap implementations (SQLite, Postgres, mock for testing)
type SettingsProvider interface {
	// Guild Settings
	SetGuildSetting(guildID, key string, value interface{}) error
	GetGuildSetting(guildID, key string, dest interface{}) error
	GetGuildSettingString(guildID, key string) (string, error)
	RemoveGuildSetting(guildID, key string) error
	ClearGuildSettings(guildID string) error

	// Global Settings
	SetGlobalSetting(key string, value interface{}) error
	GetGlobalSetting(key string, dest interface{}) error
	GetGlobalSettingString(key string) (string, error)
	RemoveGlobalSetting(key string) error

	// User Data
	SetUserData(userID, dataType string, value interface{}) error
	GetUserData(userID, dataType string, dest interface{}) error
	RemoveUserData(userID, dataType string) error

	// Helper Methods
	SetAgreementRoles(guildID string, roles []AgreementRole) error
	GetAgreementRoles(guildID string) ([]AgreementRole, error)
	SetAgreementChannel(guildID, channelID string) error
	GetAgreementChannel(guildID string) (string, error)
	SetAgreementState(userID string, state interface{}) error
	GetAgreementState(userID string, dest interface{}) error
	RemoveAgreementState(userID string) error

	// Utility
	Close() error
}

// Compile-time check that DB implements SettingsProvider
var _ SettingsProvider = (*DB)(nil)
