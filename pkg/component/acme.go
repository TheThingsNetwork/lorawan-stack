// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package component

func (c *Component) initACME() error {
	if c.config.TLS.Source != "acme" && !c.config.TLS.ACME.Enable {
		return nil
	}
	var err error
	c.acme, err = c.config.TLS.ACME.Initialize()
	if err != nil {
		return err
	}
	c.web.Prefix("/.well-known/acme-challenge/").Handler(c.acme.HTTPHandler(nil))
	return nil
}
