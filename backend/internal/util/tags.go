package util

import (
	"fmt"
	"strings"
)

const maxTagsPerLink = 10

func NormalizeTags(tags []string) ([]string, error) {
	if len(tags) > maxTagsPerLink {
		return nil, fmt.Errorf("a link can have at most %d tags", maxTagsPerLink)
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if len(tag) > 32 {
			return nil, fmt.Errorf("tags must be at most 32 characters")
		}
		for i := range tag {
			char := tag[i]
			isAlphaNumeric := char >= 'a' && char <= 'z' || char >= '0' && char <= '9'
			if !isAlphaNumeric && char != '-' && char != '_' {
				return nil, fmt.Errorf("tags may only contain letters, numbers, hyphens, and underscores")
			}
			if i == 0 && !isAlphaNumeric {
				return nil, fmt.Errorf("tags must start with a letter or number")
			}
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	if len(normalized) > maxTagsPerLink {
		return nil, fmt.Errorf("a link can have at most %d unique tags", maxTagsPerLink)
	}
	return normalized, nil
}
