// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mock

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"

var last interface{}

// Mock does nothing except save in a variable the template data of the lastest sent email.
type Mock struct{}

// New returns a new mock instance.
func New() *Mock {
	return &Mock{}
}

// Send implements email.Provider.
func (m *Mock) Send(recipient string, tmpl *templates.Template, data interface{}) error {
	last = data
	return nil
}

// Data returns the value of the template data of the lastest sent email.
func Data() interface{} {
	return last
}
