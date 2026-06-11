package service

import (
	"daily/internal/application/apperr"
	"fmt"
	"regexp"
	"strings"
)

var (
	// validTagRegex matches a valid tag format.
	validTagRegex = regexp.MustCompile(`^#[^\s\.,!?; \(\)\[\]\{\}#]+$`)
)

type TagExtractor struct{}

func NewTagExtractor() *TagExtractor {
	return &TagExtractor{}
}

// Extract scans lines from bottom to top to identify the formal tag block.
// It returns the extracted tags, the content with the tag block removed, and any error.
// It strictly expects tags to be in their own lines at the end of the content.
func (s *TagExtractor) Extract(content string) (tags []string, cleanedContent string, err error) {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, "", nil
	}

	lines := strings.Split(content, "\n")
	var allTags []string
	tagMap := make(map[string]struct{})
	lastNonTagLineIdx := -1

	// Scan from bottom to top to find the boundary of the formal tag block
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// A formal tag line MUST start with #
		if !strings.HasPrefix(line, "#") {
			lastNonTagLineIdx = i
			break
		}

		// Validate all tokens in this line as tags
		tokens := strings.Fields(line)
		var lineTags []string
		for _, token := range tokens {
			if !strings.HasPrefix(token, "#") {
				return nil, "", fmt.Errorf("%w: token in tag line must start with #: %s", apperr.ErrInvalidInput, token)
			}

			tagText := token[1:]
			if tagText == "" {
				return nil, "", fmt.Errorf("%w: tag too short: #", apperr.ErrInvalidInput)
			}
			if len(tagText) > 32 {
				return nil, "", fmt.Errorf("%w: tag too long: %s", apperr.ErrInvalidInput, token)
			}
			if !validTagRegex.MatchString(token) {
				return nil, "", fmt.Errorf("%w: invalid tag format: %s", apperr.ErrInvalidInput, token)
			}

			if _, exists := tagMap[tagText]; !exists {
				tagMap[tagText] = struct{}{}
				lineTags = append(lineTags, tagText)
			}
		}
		// Insert line tags at the beginning of allTags to maintain order
		allTags = append(lineTags, allTags...)
	}

	// Cleaned content is everything up to the last non-tag line
	if lastNonTagLineIdx == -1 {
		// Entire content is tags
		return allTags, "", nil
	}

	// Reconstruct content up to lastNonTagLineIdx
	cleanedContent = strings.Join(lines[:lastNonTagLineIdx+1], "\n")
	cleanedContent = strings.TrimSpace(cleanedContent)

	return allTags, cleanedContent, nil
}
