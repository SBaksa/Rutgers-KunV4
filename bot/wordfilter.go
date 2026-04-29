package bot

import (
	"strings"
	"unicode"
)

var slurList = []string{
	"nigger",
	"nigga",
	"faggot",
	"fag",
	"chink",
	"spic",
	"kike",
	"gook",
	"wetback",
	"tranny",
	"retard",
	"coon",
	"beaner",
	"towelhead",
	"raghead",
	"cracker",
	"dyke",
}

func containsSlur(content string) (bool, string) {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, content)

	words := strings.Fields(normalized)
	for _, word := range words {
		for _, slur := range slurList {
			if word == slur {
				return true, slur
			}
		}
	}
	return false, ""
}
