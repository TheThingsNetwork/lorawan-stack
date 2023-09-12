// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"io"

	"github.com/emersion/go-smtp"
)

type backend struct {
	messages chan *message
}

func (bkd *backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	s := &session{
		msgs: bkd.messages,
	}
	return s, nil
}

type message struct {
	Sender     string
	Recipients []string
	Data       []byte
	Opts       *smtp.MailOptions
}

type session struct {
	msg  *message
	msgs chan *message
}

func (*session) AuthPlain(_, _ string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.Reset()
	s.msg.Sender = from
	s.msg.Opts = opts
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.msg.Recipients = append(s.msg.Recipients, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.msg.Data = b
	s.msgs <- s.msg
	return nil
}

func (s *session) Reset() {
	s.msg = &message{}
}

func (*session) Logout() error {
	return nil
}
