// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import "testing"

func TestGetLogger(t *testing.T) {
	logger := GetLogger(t, "fooz")
	logger.Debug("abcabcabc - Hi!")
	logger.Info("Fooz")
	logger.Errorf("Nope %d", 1234)
}
