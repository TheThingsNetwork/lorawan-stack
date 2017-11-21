// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcclient_test

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/rpcclient"
)

func TestOptions(t *testing.T) {
	rpcclient.DefaultDialOptions(context.Background())
	// not really anything to test here
}
