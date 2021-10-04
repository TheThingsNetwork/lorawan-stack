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

// NOTE: Set CAROOT=. when running go generate, i.e.:
// $ CAROOT=. go generate .

//go:generate mkcert -cert-file servercert.pem -key-file serverkey.pem localhost 127.0.0.1 ::1
//go:generate mkcert -cert-file clientcert.pem -key-file clientkey.pem -client localhost

package testdata
