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

package cryptoutil_test

import (
	"testing"

	"github.com/smartystreets/assertions"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEmptyKeyVault(t *testing.T) {
	a := assertions.New(t)

	_, err := cryptoutil.EmptyKeyVault.Wrap(test.Context(), []byte{0x1, 0x2}, "test")
	a.So(errors.IsNotFound(err), should.BeTrue)

	_, err = cryptoutil.EmptyKeyVault.Unwrap(test.Context(), []byte{0x1, 0x2}, "test")
	a.So(errors.IsNotFound(err), should.BeTrue)

	_, err = cryptoutil.EmptyKeyVault.GetCertificate(test.Context(), "test")
	a.So(errors.IsNotFound(err), should.BeTrue)

	_, err = cryptoutil.EmptyKeyVault.ExportCertificate(test.Context(), "test")
	a.So(errors.IsNotFound(err), should.BeTrue)
}
