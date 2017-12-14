// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import "strings"

const (
	// UserPrefix is the prefix used in subjects for a user.
	UserPrefix = "user"

	// ApplicationPrefix is the prefix used in subjects for an application.
	ApplicationPrefix = "application"

	// GatewayPrefix is the prefix used in subjects for a gateway.
	GatewayPrefix = "gateway"

	// sep is the separator between the prefix and the id of the subject.
	sep = ":"
)

func splitprefix(prefix, sub string) string {
	p := prefix + sep
	if strings.HasPrefix(sub, p) {
		return strings.TrimPrefix(sub, p)
	}

	return ""
}

// UserSubject returns a subject for the user with the specified user ID.
func UserSubject(userID string) string {
	return UserPrefix + sep + userID
}

// ApplicationSubject returns a subject for the application with the specified application ID.
func ApplicationSubject(appID string) string {
	return ApplicationPrefix + sep + appID
}

// GatewaySubject returns a subject for the gateway with the specified gateway ID.
func GatewaySubject(gtwID string) string {
	return GatewayPrefix + sep + gtwID
}
