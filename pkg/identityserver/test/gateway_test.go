// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func gateway() *ttnpb.Gateway {
	return &ttnpb.Gateway{
		GatewayIdentifier: ttnpb.GatewayIdentifier{"test-gateway"},
		Description:       "My description",
		Platform:          "Kerklink",
		Attributes: map[string]string{
			"foo": "bar",
		},
		PrivacySettings: ttnpb.GatewayPrivacySettings{
			LocationPublic: true,
		},
		Antennas: []ttnpb.GatewayAntenna{
			{
				Location: ttnpb.Location{
					Latitude:  11.11,
					Longitude: 22.22,
					Altitude:  10,
				},
			},
		},
	}
}

func TestShouldBeGateway(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeGateway(gateway(), gateway()), should.Equal, success)

	modified := gateway()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeGateway(modified, gateway()), should.NotEqual, success)
}

func TestShouldBeGatewayIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeGatewayIgnoringAutoFields(gateway(), gateway()), should.Equal, success)

	modified := gateway()
	modified.Platform = "foo"
	modified.Attributes["foz"] = "baz"
	modified.Antennas = append(modified.Antennas, ttnpb.GatewayAntenna{
		Gain: 12.12,
	})

	a.So(ShouldBeGatewayIgnoringAutoFields(modified, gateway()), should.NotEqual, success)
}
