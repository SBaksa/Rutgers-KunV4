package config

import (
	"os"
	"strings"
)

func IsOwner(userID string) bool {
	owners := os.Getenv("BOT_OWNERS")
	if owners == "" {
		return false
	}
	for _, id := range strings.Split(owners, ",") {
		if strings.TrimSpace(id) == userID {
			return true
		}
	}
	return false
}
