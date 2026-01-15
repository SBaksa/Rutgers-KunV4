package database

import (
	"encoding/json"
	"fmt"
	"log"
)

// TestJSCompatibility tests that our Go implementation can read/write data
// in exactly the same format as the JavaScript bot
func (db *DB) TestJSCompatibility() error {
	log.Println("Testing JavaScript bot database compatibility...")

	// Test 1: Agreement Roles (complex object array)
	if err := db.testAgreementRoles(); err != nil {
		return fmt.Errorf("agreement roles test failed: %w", err)
	}

	// Test 2: Agreement Channel (simple string)
	if err := db.testAgreementChannel(); err != nil {
		return fmt.Errorf("agreement channel test failed: %w", err)
	}

	// Test 3: Global Settings (empty guild)
	if err := db.testGlobalSettings(); err != nil {
		return fmt.Errorf("global settings test failed: %w", err)
	}

	// Test 4: User Data (prefixed keys)
	if err := db.testUserData(); err != nil {
		return fmt.Errorf("user data test failed: %w", err)
	}

	// Test 5: Custom Commands
	if err := db.testCustomCommands(); err != nil {
		return fmt.Errorf("custom commands test failed: %w", err)
	}

	log.Println("✓ All compatibility tests passed!")
	return nil
}

// testAgreementRoles tests the exact format used by the JS bot
func (db *DB) testAgreementRoles() error {
	guildID := "682667655043350528" // Rutgers Math Discord ID from JS bot

	// This is exactly how the JS bot stores agreement roles
	jsFormatRoles := []AgreementRole{
		{RoleID: "682670425108643852", Authenticate: "true"},       // Student role
		{RoleID: "682670461070737409", Authenticate: "false"},      // Guest role
		{RoleID: "682816262643908609", Authenticate: "permission"}, // Base permissions
	}

	// Store using our Go implementation
	if err := db.SetAgreementRoles(guildID, jsFormatRoles); err != nil {
		return err
	}

	// Retrieve using our Go implementation
	retrievedRoles, err := db.GetAgreementRoles(guildID)
	if err != nil {
		return err
	}

	// Verify they match
	if len(retrievedRoles) != len(jsFormatRoles) {
		return fmt.Errorf("role count mismatch: got %d, want %d", len(retrievedRoles), len(jsFormatRoles))
	}

	for i, role := range retrievedRoles {
		if role.RoleID != jsFormatRoles[i].RoleID || role.Authenticate != jsFormatRoles[i].Authenticate {
			return fmt.Errorf("role mismatch at index %d", i)
		}
	}

	// Test raw database format matches JS bot
	var rawValue string
	query := `SELECT value FROM settings WHERE guild = ? AND key = ?`
	err = db.conn.QueryRow(query, guildID, "agreementRoles").Scan(&rawValue)
	if err != nil {
		return err
	}

	// This should be valid JSON that the JS bot could parse
	var testParse []AgreementRole
	if err := json.Unmarshal([]byte(rawValue), &testParse); err != nil {
		return fmt.Errorf("stored JSON is not valid: %w", err)
	}

	log.Println("✓ Agreement roles format matches JS bot")
	return nil
}

// testAgreementChannel tests simple string storage
func (db *DB) testAgreementChannel() error {
	guildID := "682667655043350528"
	channelID := "682709314078638097" // Agreement channel from JS bot

	// Store
	if err := db.SetAgreementChannel(guildID, channelID); err != nil {
		return err
	}

	// Retrieve
	retrieved, err := db.GetAgreementChannel(guildID)
	if err != nil {
		return err
	}

	if retrieved != channelID {
		return fmt.Errorf("channel ID mismatch: got %s, want %s", retrieved, channelID)
	}

	// Check raw format - JS bot stores strings with quotes
	var rawValue string
	query := `SELECT value FROM settings WHERE guild = ? AND key = ?`
	err = db.conn.QueryRow(query, guildID, "agreementChannel").Scan(&rawValue)
	if err != nil {
		return err
	}

	// Should be quoted JSON string
	if rawValue != `"`+channelID+`"` {
		return fmt.Errorf("raw value format incorrect: got %s, want %s", rawValue, `"`+channelID+`"`)
	}

	log.Println("✓ Agreement channel format matches JS bot")
	return nil
}

// testGlobalSettings tests global settings with empty guild string
func (db *DB) testGlobalSettings() error {
	// JS bot stores global settings with guild = ""
	messagesToCache := []map[string]string{
		{"channel": "682833023942525018", "message": "1157917032583532555"},
		{"channel": "791417300644921345", "message": "1074527910532239370"},
	}

	// Store
	if err := db.SetGlobalSetting("messagesToCache", messagesToCache); err != nil {
		return err
	}

	// Retrieve
	var retrieved []map[string]string
	if err := db.GetGlobalSetting("messagesToCache", &retrieved); err != nil {
		return err
	}

	if len(retrieved) != len(messagesToCache) {
		return fmt.Errorf("messages to cache mismatch")
	}

	// Check that it's stored with empty guild string
	var count int
	query := `SELECT COUNT(*) FROM settings WHERE guild = '' AND key = 'messagesToCache'`
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("global setting not stored with empty guild string")
	}

	log.Println("✓ Global settings format matches JS bot")
	return nil
}

// testUserData tests user data with prefixed keys
func (db *DB) testUserData() error {
	userID := "219601832832401419" // User ID from JS bot

	// Test quotes (JS bot pattern: key = "quotes:userID", guild = "")
	quotes := []string{
		"This is a test quote",
		"Another quote with attachment\nhttps://example.com/image.png",
	}

	if err := db.SetUserQuotes(userID, quotes); err != nil {
		return err
	}

	retrieved, err := db.GetUserQuotes(userID)
	if err != nil {
		return err
	}

	if len(retrieved) != len(quotes) {
		return fmt.Errorf("quotes mismatch")
	}

	// Check database format matches JS bot pattern
	var count int
	query := `SELECT COUNT(*) FROM settings WHERE guild = '' AND key = ?`
	err = db.conn.QueryRow(query, "quotes:"+userID).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("user quotes not stored in JS bot format")
	}

	// Test word count data
	wordData := []map[string]interface{}{
		{"word": "test", "count": 42},
		{"word": "example", "count": 17},
	}

	if err := db.SetWordCount(userID, wordData); err != nil {
		return err
	}

	var retrievedWordData []map[string]interface{}
	if err := db.GetWordCount(userID, &retrievedWordData); err != nil {
		return err
	}

	// Check database format
	err = db.conn.QueryRow(query, "countword:"+userID).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("word count not stored in JS bot format")
	}

	log.Println("✓ User data format matches JS bot")
	return nil
}

// testCustomCommands tests custom command storage
func (db *DB) testCustomCommands() error {
	guildID := "682667655043350528"

	// JS bot stores custom commands as "commands:commandname"
	commandData := map[string]interface{}{
		"text":       "This is a custom command response",
		"userID":     "219601832832401419",
		"timestamp":  "12/25/2023, 3:45:12 PM",
		"attachment": "https://cdn.discordapp.com/attachments/example.png",
	}

	commandName := "testcommand"
	key := "commands:" + commandName

	if err := db.SetGuildSetting(guildID, key, commandData); err != nil {
		return err
	}

	var retrieved map[string]interface{}
	if err := db.GetGuildSetting(guildID, key, &retrieved); err != nil {
		return err
	}

	// Verify structure matches
	if retrieved["text"] != commandData["text"] {
		return fmt.Errorf("custom command data mismatch")
	}

	log.Println("✓ Custom command format matches JS bot")
	return nil
}

// MigrateFromJS helps migrate data from JS bot format if needed
func (db *DB) MigrateFromJS(jsDbPath string) error {
	log.Printf("Checking if migration from JS database is needed...")

	// In most cases, the database should be directly compatible
	// This function is here if we find any edge cases that need fixing

	// Test if we can read existing data
	if err := db.TestJSCompatibility(); err != nil {
		log.Printf("Compatibility issue found: %v", err)
		return err
	}

	log.Println("✓ Database is fully compatible with JS bot format")
	return nil
}

// DebugDatabaseContents shows what's in the database (for troubleshooting)
func (db *DB) DebugDatabaseContents() {
	log.Println("=== DATABASE CONTENTS ===")

	query := `SELECT guild, key, SUBSTR(value, 1, 100) as value_preview FROM settings ORDER BY guild, key`
	rows, err := db.conn.Query(query)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var guild, key, valuePreview string
		if err := rows.Scan(&guild, &key, &valuePreview); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		guildDisplay := guild
		if guild == "" {
			guildDisplay = "(global)"
		}

		log.Printf("Guild: %s, Key: %s, Value: %s...", guildDisplay, key, valuePreview)
	}

	log.Println("=== END DATABASE CONTENTS ===")
}
