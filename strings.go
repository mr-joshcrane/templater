package templater

import (
	"path/filepath"
	"regexp"
	"strings"
)

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

var validCharacters = regexp.MustCompile(`[A-Z0-9._ ]*`)
var camelCase = regexp.MustCompile(`([a-z])(A?)([A-Z])`)
func NormaliseKey(s string) string {
	s = camelCase.ReplaceAllString(s, `$1 $2 $3`)
	s = strings.ToUpper(s)
	s = strings.Join(validCharacters.FindAllString(s, -1), " ")
	s = strings.Join(strings.Fields(s), "_")
	s = strings.Trim(s, ` `)
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
