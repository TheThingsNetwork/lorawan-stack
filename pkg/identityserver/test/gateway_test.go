// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
		FrequencyPlanID:   "868_3",
		Token:             "1111",
		Platform:          "Kerklink",
		ClusterAddress:    "localhost",
		Attributes: map[string]string{
			"foo": "bar",
		},
		PrivacySettings: ttnpb.GatewayPrivacySettings{
			LocationPublic: true,
		},
		Antennas: []ttnpb.GatewayAntenna{
			ttnpb.GatewayAntenna{
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
	modified.Token = "foo"
	modified.Attributes["foz"] = "baz"

	a.So(ShouldBeGatewayIgnoringAutoFields(modified, gateway()), should.NotEqual, success)
}
