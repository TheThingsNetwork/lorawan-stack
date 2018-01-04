// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package util

import (
	"strings"

	"github.com/gobwas/glob"
)

// IsIDAllowed checks whether an ID is allowed to be used given the list of
// blacklisted IDs contained in the settings.
func IsIDAllowed(id string, blacklistedIDs []string) bool {
	if blacklistedIDs == nil || len(blacklistedIDs) == 0 {
		return true
	}

	allowed := true
	for _, blacklistedID := range blacklistedIDs {
		allowed = id != blacklistedID
		if !allowed {
			break
		}
	}

	return allowed
}

// IsEmailAllowed checks whether an input email is allowed given the glob list
// of allowed emails in the settings.
func IsEmailAllowed(email string, allowedEmails []string) bool {
	if allowedEmails == nil || len(allowedEmails) == 0 {
		return true
	}

	found := false
	for i := range allowedEmails {
		found = glob.MustCompile(strings.ToLower(allowedEmails[i])).Match(strings.ToLower(email))
		if found {
			break
		}
	}

	return found
}
