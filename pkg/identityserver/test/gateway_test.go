// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func gateway() *types.DefaultGateway {
	return &types.DefaultGateway{
		ID:            "test-gateway",
		Description:   "My description",
		FrequencyPlan: "868_3",
		Key:           "1111",
		Brand:         utils.StringAddress("Kerklink"),
		Routers:       []string{"network.eu", "network.au"},
		Attributes: map[string]string{
			"foo": "bar",
		},
		Antennas: []types.GatewayAntenna{
			types.GatewayAntenna{
				ID: "test antenna",
				Location: &ttnpb.Location{
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
	modified.Created = time.Now()

	a.So(ShouldBeGateway(modified, gateway()), should.NotEqual, success)
}

func TestShouldBeGatewayIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeGatewayIgnoringAutoFields(gateway(), gateway()), should.Equal, success)

	modified := gateway()
	modified.Key = "foo"
	modified.Attributes["foz"] = "baz"

	a.So(ShouldBeGatewayIgnoringAutoFields(modified, gateway()), should.NotEqual, success)
}
