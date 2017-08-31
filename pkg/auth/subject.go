// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import "strings"

const (
	applicationPrefix = "application:"
	gatewayPrefix     = "gateway:"
	userPrefix        = "user:"
)

// Subject is the type of subjects for a JWT token.
type Subject string

// String implements fmt.Stringer.
func (s Subject) String() string {
	return string(s)
}

// ApplicationSubject returns a JWT subject that targets the application specified by the provided application ID.
func ApplicationSubject(appID string) Subject {
	return Subject(applicationPrefix + appID)
}

// GatewaySubject returns a JWT subject that targets the gateway specified by the provided gateway ID.
func GatewaySubject(gtwID string) Subject {
	return Subject(gatewayPrefix + gtwID)
}

// UserSubject returns a JWT subject that targets the user specified by the provided username.
func UserSubject(username string) Subject {
	return Subject(userPrefix + username)
}

func (s Subject) match(prefix string) string {
	str := s.String()
	if strings.HasPrefix(str, prefix) {
		return strings.TrimPrefix(str, prefix)
	}
	return ""
}

// Application returns the application ID of the applciation the subject is for, or the empty string if it is not for an application.
func (s Subject) Application() string {
	return s.match(applicationPrefix)
}

// Gateway returns the gateway ID of the gateway the subject is for, or the empty string if it is not for a gateway.
func (s Subject) Gateway() string {
	return s.match(gatewayPrefix)
}

// User returns the username of the user the subject is for, or the empty string if it is not for a user.
func (s Subject) User() string {
	return s.match(userPrefix)
}
