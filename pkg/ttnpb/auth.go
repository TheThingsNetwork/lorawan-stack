// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// Username returns the username of the user this scope is for, or the empty string if it is not for a user.
func (s Scope) Username() string {
	if s.Type == SCOPE_USER {
		return s.ID
	}
	return ""
}

// ApplicationID returns the application ID of the application this scope is for, or the empty string if it is not for an application.
func (s Scope) ApplicationID() string {
	if s.Type == SCOPE_APPLICATION {
		return s.ID
	}
	return ""
}

// GatewayID returns the gateway ID of the gateway this scope is for, or the empty string if it is not for a gateway.
func (s Scope) GatewayID() string {
	if s.Type == SCOPE_GATEWAY {
		return s.ID
	}
	return ""
}

// hasRight checks wether or not the right is included in this scope.
func (s Scope) hasRight(right Right) bool {
	for _, r := range s.Rights {
		if r == right {
			return true
		}
	}
	return false
}

// HasRights checks wether or not the provided right is included in the scope. It will only return true if all the provided rights are
// included in the token..
func (s Scope) HasRights(rights ...Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && s.hasRight(right)
	}

	return ok
}

// UserScope returns a scope with the specified rights that is valid for the specified user.
func UserScope(username string, rights ...Right) Scope {
	return Scope{
		Type:   SCOPE_USER,
		ID:     username,
		Rights: rights,
	}
}

// ApplicationScope returns a scope with the specified rights that is valid for the specified application.
func ApplicationScope(applicationID string, rights ...Right) Scope {
	return Scope{
		Type:   SCOPE_APPLICATION,
		ID:     applicationID,
		Rights: rights,
	}
}

// GatewayScopereturns a scope with the specified rights that is valid for the specified gateway.
func GatewayScope(gatewayID string, rights ...Right) Scope {
	return Scope{
		Type:   SCOPE_GATEWAY,
		ID:     gatewayID,
		Rights: rights,
	}
}

// ClientScope a scope with the specified rights that is valid for the specified client.
func ClientScope(gatewayID string, rights ...Right) Scope {
	return Scope{
		Type:   SCOPE_GATEWAY,
		ID:     gatewayID,
		Rights: rights,
	}
}
