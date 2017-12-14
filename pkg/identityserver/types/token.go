// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// ValidationToken is an expirable token.
type ValidationToken struct {
	// ValidationToken is the token itself.
	ValidationToken string

	// CreatedAt denotes when the token was created.
	CreatedAt time.Time

	// ExpiresIn denotes the TTL of the token in seconds.
	ExpiresIn int32
}

// IsExpired checks whether the token is expired or not.
func (v ValidationToken) IsExpired() bool {
	return v.CreatedAt.Add(time.Duration(v.ExpiresIn) * time.Second).Before(time.Now())
}
