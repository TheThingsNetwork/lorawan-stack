package validation

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/udp"
)

func TestAlwaysValid(t *testing.T) {
	v := AlwaysValid()
	p := udp.Packet{}

	if v.Valid(p) == false {
		t.Error("AlwaysValid should return true")
	}
}
