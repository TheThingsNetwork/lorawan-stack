// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mock

import "go.thethings.network/lorawan-stack/pkg/identityserver/email/templates"

var last interface{}

// Mock does nothing except save in a variable the template data of the lastest sent email.
type Mock struct{}

// New returns a new mock instance.
func New() *Mock {
	return &Mock{}
}

// Send implements email.Provider.
// It pushs the data value to a variable that can be accessed through the Data method of this package.
func (m *Mock) Send(recipient string, template templates.Template) error {
	last = template
	return nil
}

// Data pops the value of the template data of the lastest sent email.
func Data() interface{} {
	defer func() { last = nil }()
	return last
}
