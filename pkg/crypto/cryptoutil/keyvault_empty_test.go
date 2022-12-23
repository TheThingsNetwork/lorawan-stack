// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestEmptyKeyVault(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	_, err := cryptoutil.EmptyKeyVault.Key(test.Context(), "test")
	a.So(errors.IsNotFound(err), should.BeTrue)

	_, err = cryptoutil.EmptyKeyVault.ServerCertificate(test.Context(), "test")
	a.So(errors.IsNotFound(err), should.BeTrue)

	_, err = cryptoutil.EmptyKeyVault.ClientCertificate(test.Context())
	a.So(errors.IsNotFound(err), should.BeTrue)
}
