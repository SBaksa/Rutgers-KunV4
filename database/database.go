package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

var Instance *DB

// Initialize sets up the SQLite database with JS bot compatible schema
func Initialize(dbPath string) error {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	Instance = &DB{conn: conn}

	if err := Instance.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables sets up the database schema exactly like the JS bot
func (db *DB) createTables() error {
	// Single settings table - matches Discord.js Commando SQLiteProvider
	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild TEXT NOT NULL DEFAULT '',
		key TEXT NOT NULL,
		value TEXT,
		UNIQUE(guild, key)
	);`

	if _, err := db.conn.Exec(settingsTable); err != nil {
		return fmt.Errorf("failed to create settings table: %w", err)
	}

	return nil
}

// Guild Settings Methods

// SetGuildSetting stores a guild-specific setting
func (db *DB) SetGuildSetting(guildID, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	query := `INSERT OR REPLACE INTO settings (guild, key, value) VALUES (?, ?, ?)`
	_, err = db.conn.Exec(query, guildID, key, string(jsonValue))
	return err
}

// GetGuildSetting retrieves a guild-specific setting
func (db *DB) GetGuildSetting(guildID, key string, dest interface{}) error {
	query := `SELECT value FROM settings WHERE guild = ? AND key = ?`
	var jsonValue string

	err := db.conn.QueryRow(query, guildID, key).Scan(&jsonValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("setting not found")
		}
		return err
	}

	return json.Unmarshal([]byte(jsonValue), dest)
}

// GetGuildSettingString is a helper for string values
func (db *DB) GetGuildSettingString(guildID, key string) (string, error) {
	var value string
	err := db.GetGuildSetting(guildID, key, &value)
	return value, err
}

// RemoveGuildSetting deletes a guild-specific setting
func (db *DB) RemoveGuildSetting(guildID, key string) error {
	query := `DELETE FROM settings WHERE guild = ? AND key = ?`
	_, err := db.conn.Exec(query, guildID, key)
	return err
}

// ClearGuildSettings removes all settings for a guild
func (db *DB) ClearGuildSettings(guildID string) error {
	query := `DELETE FROM settings WHERE guild = ?`
	_, err := db.conn.Exec(query, guildID)
	return err
}

// Global Settings Methods (use empty guild string like JS bot)

// SetGlobalSetting stores a global setting using empty guild string
func (db *DB) SetGlobalSetting(key string, value interface{}) error {
	return db.SetGuildSetting("", key, value)
}

// GetGlobalSetting retrieves a global setting
func (db *DB) GetGlobalSetting(key string, dest interface{}) error {
	return db.GetGuildSetting("", key, dest)
}

// GetGlobalSettingString is a helper for string values
func (db *DB) GetGlobalSettingString(key string) (string, error) {
	return db.GetGuildSettingString("", key)
}

// RemoveGlobalSetting deletes a global setting
func (db *DB) RemoveGlobalSetting(key string) error {
	return db.RemoveGuildSetting("", key)
}

// User Data Methods (using the JS bot pattern of key prefixes)

// SetUserData stores user-specific data using prefixed keys like "quotes:userID"
func (db *DB) SetUserData(userID, dataType string, value interface{}) error {
	key := fmt.Sprintf("%s:%s", dataType, userID)
	return db.SetGlobalSetting(key, value)
}

// GetUserData retrieves user-specific data
func (db *DB) GetUserData(userID, dataType string, dest interface{}) error {
	key := fmt.Sprintf("%s:%s", dataType, userID)
	return db.GetGlobalSetting(key, dest)
}

// RemoveUserData deletes user-specific data
func (db *DB) RemoveUserData(userID, dataType string) error {
	key := fmt.Sprintf("%s:%s", dataType, userID)
	return db.RemoveGlobalSetting(key)
}

// Utility Methods

// GetAllGuildSettings returns all settings for a guild (for debugging)
func (db *DB) GetAllGuildSettings(guildID string) (map[string]string, error) {
	query := `SELECT key, value FROM settings WHERE guild = ?`
	rows, err := db.conn.Query(query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	return settings, nil
}

// GetAllSettings returns all settings (for debugging/migration)
func (db *DB) GetAllSettings() (map[string]map[string]string, error) {
	query := `SELECT guild, key, value FROM settings`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]map[string]string)
	for rows.Next() {
		var guild, key, value string
		if err := rows.Scan(&guild, &key, &value); err != nil {
			return nil, err
		}

		if settings[guild] == nil {
			settings[guild] = make(map[string]string)
		}
		settings[guild][key] = value
	}

	return settings, nil
}

// RawQuery allows direct SQL queries (for debugging/migration)
func (db *DB) RawQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// RawExec allows direct SQL execution (for debugging/migration)
func (db *DB) RawExec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Agreement Role Types (for type safety)
type AgreementRole struct {
	RoleID       string `json:"roleID"`
	Authenticate string `json:"authenticate"` // "true", "false", or "permission"
}

type AgreementSlim struct {
	Message string `json:"message"`
	Emote   string `json:"emote"`
	Role    string `json:"role"`
}

// Helper functions for common settings

// SetAgreementRoles sets the agreement roles for a guild
func (db *DB) SetAgreementRoles(guildID string, roles []AgreementRole) error {
	return db.SetGuildSetting(guildID, "agreementRoles", roles)
}

// GetAgreementRoles gets the agreement roles for a guild
func (db *DB) GetAgreementRoles(guildID string) ([]AgreementRole, error) {
	var roles []AgreementRole
	err := db.GetGuildSetting(guildID, "agreementRoles", &roles)
	return roles, err
}

// SetAgreementChannel sets the agreement channel for a guild
func (db *DB) SetAgreementChannel(guildID, channelID string) error {
	return db.SetGuildSetting(guildID, "agreementChannel", channelID)
}

// GetAgreementChannel gets the agreement channel for a guild
func (db *DB) GetAgreementChannel(guildID string) (string, error) {
	return db.GetGuildSettingString(guildID, "agreementChannel")
}

// SetUserQuotes sets quotes for a user (matches JS bot pattern exactly)
func (db *DB) SetUserQuotes(userID string, quotes []string) error {
	return db.SetUserData(userID, "quotes", quotes)
}

// GetUserQuotes gets quotes for a user
func (db *DB) GetUserQuotes(userID string) ([]string, error) {
	var quotes []string
	err := db.GetUserData(userID, "quotes", &quotes)
	if err != nil {
		// Return empty slice if not found (matches JS bot behavior)
		return []string{}, nil
	}
	return quotes, nil
}

// SetWordCount sets word count data for a user (matches JS bot pattern exactly)
func (db *DB) SetWordCount(userID string, wordData interface{}) error {
	return db.SetUserData(userID, "countword", wordData)
}

// GetWordCount gets word count data for a user
func (db *DB) GetWordCount(userID string, dest interface{}) error {
	return db.GetUserData(userID, "countword", dest)
}

// SetCustomCommand stores a custom command (matches JS bot pattern)
func (db *DB) SetCustomCommand(guildID, commandName string, commandData interface{}) error {
	key := fmt.Sprintf("commands:%s", commandName)
	return db.SetGuildSetting(guildID, key, commandData)
}

// GetCustomCommand retrieves a custom command
func (db *DB) GetCustomCommand(guildID, commandName string, dest interface{}) error {
	key := fmt.Sprintf("commands:%s", commandName)
	return db.GetGuildSetting(guildID, key, dest)
}

// RemoveCustomCommand deletes a custom command
func (db *DB) RemoveCustomCommand(guildID, commandName string) error {
	key := fmt.Sprintf("commands:%s", commandName)
	return db.RemoveGuildSetting(guildID, key)
}

// GetAllCustomCommands returns all custom command names for a guild
func (db *DB) GetAllCustomCommands(guildID string) ([]string, error) {
	query := `SELECT key FROM settings WHERE guild = ? AND key LIKE 'commands:%'`
	rows, err := db.conn.Query(query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		// Extract command name from "commands:commandname"
		commandName := strings.TrimPrefix(key, "commands:")
		commands = append(commands, commandName)
	}

	return commands, nil
}

// SetAgreementState stores user agreement state (matches JS bot pattern)
func (db *DB) SetAgreementState(userID string, state interface{}) error {
	key := fmt.Sprintf("agree:%s", userID)
	return db.SetGlobalSetting(key, state)
}

// GetAgreementState retrieves user agreement state
func (db *DB) GetAgreementState(userID string, dest interface{}) error {
	key := fmt.Sprintf("agree:%s", userID)
	return db.GetGlobalSetting(key, dest)
}

// RemoveAgreementState removes user agreement state
func (db *DB) RemoveAgreementState(userID string) error {
	key := fmt.Sprintf("agree:%s", userID)
	return db.RemoveGlobalSetting(key)
}
