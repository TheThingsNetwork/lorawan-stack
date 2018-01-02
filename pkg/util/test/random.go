// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"math/rand"

	"github.com/TheThingsNetwork/ttn/pkg/util/randutil"
)

// Randy is global rand safe for concurrent use
var Randy = rand.New(randutil.NewLockedSource(rand.NewSource(42)))
