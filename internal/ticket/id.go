package ticket

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// GenerateID generates a new ticket ID based on the current directory name
// and a 4-character nanoid using cryptographic randomness for uniqueness.
// Format: {prefix}-{4-char-alphanumeric}
// The prefix is extracted from the directory name by taking the first letter
// of each hyphen/underscore-separated segment, or the first 3 chars as fallback.
// The nanoid uses lowercase alphanumeric characters (a-z0-9), providing
// 36^4 = 1,679,616 possible IDs per prefix.
func GenerateID(cwd string) string {
	dirName := filepath.Base(cwd)

	// Extract first letter of each segment (split by hyphen or underscore)
	segments := strings.FieldsFunc(dirName, func(r rune) bool {
		return r == '-' || r == '_'
	})

	var prefix string
	for _, seg := range segments {
		if len(seg) > 0 {
			// Take first rune to handle unicode correctly
			for _, r := range seg {
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					prefix += string(r)
					break
				}
			}
		}
	}

	// Fallback to first 3 chars if no segments produced a prefix
	if prefix == "" {
		runes := []rune(dirName)
		if len(runes) > 3 {
			prefix = string(runes[:3])
		} else {
			prefix = dirName
		}
	}

	// 4-char nanoid with lowercase alphanumeric charset (a-z0-9)
	alphabet := "abcdefghijklmnopqrstuvwxyz0123456789"
	hashStr, err := gonanoid.Generate(alphabet, 4)
	if err != nil {
		// Fallback to timestamp-based (extremely unlikely)
		hashStr = fmt.Sprintf("%04d", time.Now().UnixNano()%10000)
	}

	return fmt.Sprintf("%s-%s", strings.ToLower(prefix), hashStr)
}
