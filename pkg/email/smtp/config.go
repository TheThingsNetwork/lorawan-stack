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

package smtp

import (
	"crypto/tls"
	"net"
	"net/smtp"
)

// Config for the SMTP email provider.
type Config struct {
	Address     string `name:"address" description:"SMTP server address"`
	Username    string `name:"username" description:"Username to authenticate with"`
	Password    string `name:"password" description:"Password to authenticate with"`
	Connections int    `name:"connections" description:"Maximum number of connections to the SMTP server"`
	TLSConfig   *tls.Config
}

func (c Config) auth() smtp.Auth {
	if c.Username == "" && c.Password == "" {
		return nil
	}
	host, _, _ := net.SplitHostPort(c.Address)
	return smtp.PlainAuth("", c.Username, c.Password, host)
}
