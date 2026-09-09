package adsefid

import (
	"regexp"
	"strings"
)

// localIDPattern matches a valid local_id: 1-36 ASCII letters/digits, with
// '-', '_', '.', ':' allowed only strictly between the first and last
// character.
var localIDPattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9\-_.:]{0,34}[A-Za-z0-9])?$`)

const (
	maxSmsMessageLength       = 900
	maxMessengerMessageLength = 4000
	maxCombinedIDsLookup      = 2000
	minReceivedCount          = 1
	maxReceivedCount          = 499
	minTemplatesTake          = 1
	maxTemplatesTake          = 100
)

// validateLocalID mirrors the service, which normalizes a blank local_id to
// "not supplied" before validating it. A nil, empty or whitespace-only value is
// therefore accepted and simply omitted from the request.
func validateLocalID(localID *string, field string) *ValidationError {
	if localID == nil || strings.TrimSpace(*localID) == "" {
		return nil
	}
	if !localIDPattern.MatchString(*localID) {
		return &ValidationError{
			Field:   field,
			Message: "must be 1-36 ASCII letters/digits, with '-', '_', '.', ':' allowed only between the first and last character",
		}
	}
	return nil
}

func requireNonEmpty(value, field string) *ValidationError {
	if value == "" {
		return &ValidationError{Field: field, Message: "is required"}
	}
	return nil
}

func requireRequest[T any](value *T) *ValidationError {
	if value == nil {
		return &ValidationError{Field: "request", Message: "is required"}
	}
	return nil
}

// requireMaxLength counts UTF-16 code units, not runes or bytes, because that
// is what the server counts. A character outside the Basic Multilingual Plane
// (an emoji, say) is one rune but two UTF-16 code units, so counting runes here
// would accept a message the server rejects.
func requireMaxLength(value string, maxLength int, field string) *ValidationError {
	if utf16Length(value) > maxLength {
		return &ValidationError{Field: field, Message: "exceeds the maximum allowed length"}
	}
	return nil
}

func utf16Length(value string) int {
	length := 0
	for _, r := range value {
		if r > 0xFFFF {
			length += 2
		} else {
			length++
		}
	}
	return length
}

func requireNonEmptySlice[T any](values []T, field string) *ValidationError {
	if len(values) == 0 {
		return &ValidationError{Field: field, Message: "must contain at least one item"}
	}
	return nil
}

func requireInRange(value, minValue, maxValue int, field string) *ValidationError {
	if value < minValue || value > maxValue {
		return &ValidationError{Field: field, Message: "is out of the allowed range"}
	}
	return nil
}

func requireAtLeastOne(firstProvided, secondProvided bool, message string) *ValidationError {
	if !firstProvided && !secondProvided {
		return &ValidationError{Message: message}
	}
	return nil
}

func requireCombinedCountAtMost(count1, count2, maxAllowed int, message string) *ValidationError {
	if count1+count2 > maxAllowed {
		return &ValidationError{Message: message}
	}
	return nil
}

// validateIDsLookup enforces the shared rules of the get-status endpoints: at
// least one id list, and at most maxCombinedIDsLookup distinct ids in total.
func validateIDsLookup(messageIDs, localIDs []string) *ValidationError {
	if err := requireAtLeastOne(len(messageIDs) > 0, len(localIDs) > 0, "at least one of MessageIDs or LocalIDs is required"); err != nil {
		return err
	}
	return requireCombinedCountAtMost(distinctCount(messageIDs), distinctCount(localIDs), maxCombinedIDsLookup, "the combined distinct count of MessageIDs and LocalIDs must not exceed 2000")
}

func distinctCount(values []string) int {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	return len(seen)
}
