package unpack

import (
	"errors"
	"strings"
	"unicode"
)

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var result strings.Builder
	var prev rune
	var hasPrev bool
	escaped := false

	for _, r := range s {
		if escaped {
			prev = r
			hasPrev = true
			result.WriteRune(prev)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if unicode.IsDigit(r) {
			if !hasPrev {
				return "", errors.New("invalid string")
			}

			count := int(r - '0')

			if count == 0 {
				hasPrev = false
				continue
			}

			for i := 1; i < count; i++ {
				result.WriteRune(prev)
			}

			hasPrev = false
			continue
		}

		prev = r
		hasPrev = true
		result.WriteRune(prev)
	}

	if escaped {
		return "", errors.New("invalid string")
	}

	return result.String(), nil
}
