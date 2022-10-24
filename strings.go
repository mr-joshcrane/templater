package templater

import (
	"path/filepath"
	"regexp"
	"strings"
)

var arrayAtLineStart = regexp.MustCompile(`^[[0-9]*].`)
var arrayInLine = regexp.MustCompile(`[\[[0-9]]`)

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == '_' ||
			b == ' ' ||
			b == '.' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func formatKey(s string) string {
	s = strings.ToUpper(s)
	s = strings.ReplaceAll(s, `(`, " ")
	s = strip(s)
	s = strings.TrimLeft(s, ` `)
	s = strings.ReplaceAll(s, `.`, `__`)
	s = strings.ReplaceAll(s, ` `, `_`)
	return s
}

func cleanTableName(path string) string {
	tableName := filepath.Base(path)
	tableName = strings.ToUpper(tableName)
	tableName = strings.ReplaceAll(tableName, ".CSV", "")
	return tableName
}

type NameOption func(string) string

func prefix(s string) string {
	return "V:" + s
}

func stripInitialArray(s string) string {
	return arrayAtLineStart.ReplaceAllString(s, "")
}

func stripAndEscapeQuotes(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `:`, `":"`)
	s = strings.ReplaceAll(s, `.`, `"."`)
	return s
}

func containsArray(s string) bool {
	return arrayInLine.MatchString(s)
}
