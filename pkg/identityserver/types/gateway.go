// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// RouterList is the type that represent a list of routers a gateway connect to
type RouterList []string

// GatewayAntennaPlacement denotes whether a gateway antenna is placed indoors or outdoors
type GatewayAntennaPlacement string

const (
	// PlacementIndoor means that the antenna is placed indoors
	PlacementIndoor GatewayAntennaPlacement = "indoors"

	// PlacementOutdoor means that the antenna is placed outdoors
	PlacementOutdoor GatewayAntennaPlacement = "outdoors"
)

// DefaultGateway represents an gateway on the network
type DefaultGateway struct {
	// ID is the unique id of the application
	ID string `db:"id"`

	// Description is the description of the application
	Description string `db:"description"`

	// FrequencyPlan is the name of the frequency plan the gateway is using
	FrequencyPlan string `db:"frequency_plan"`

	// Key is the secret key the gateway uses to prove its identity
	Key string `db:"key"`

	// Activated denotes wether or not the gateway has been activated
	Activated bool `db:"activated"`

	// StatusPublic denotes whether or not the gateway's status is public or not
	StatusPublic bool `db:"status_public"`

	// LocationPublic denotes whether or not the gateway's location is public
	LocationPublic bool `db:"location_public"`

	// OwnerPublic denotes whether or not the gateway owner is public
	OwnerPublic bool `db:"owner_public"`

	// AutoUpdate indicates whether or not the gateway should be able to
	// automatically fetch and execute firmware updates
	AutoUpdate bool `db:"auto_update"`

	// Brand is the gateway brand
	Brand *string `db:"brand"`

	// Model is the gateway model
	Model *string `db:"model"`

	// Antennas is all the antennas that the gateway has
	Antennas []GatewayAntenna `json:"antennas"`

	// Attributes is a free-form of attributes
	Attributes map[string]string

	// Routers is a list of router id's that will be tried in order
	// Internally they are stored in the same column separated by comma
	Routers RouterList `db:"routers"`

	// Created is the time when the gateway was created
	Created time.Time `json:"created" db:"created"`

	// Deleted is the time when the user account was deleted
	Archived *time.Time `json:"deleted,omitempty" db:"archived"`
}

// GatewayAntenna is a gateway antenna
type GatewayAntenna struct {
	// ID is the unique and immutable antenna's identifier
	ID string `db:"antenna_id"`

	// Location is the antenna's location defined by: latitude, longitude and altitude
	Location *ttnpb.Location

	// Type denotes the antenna's type
	Type *string `db:"type"`

	// Model denotes the antenna's model
	Model *string `db:"model"`

	// Placement denotes whether if the antenna is placed indoors or outdoors
	Placement *GatewayAntennaPlacement `db:"placement"`
}

// Gateway is the interface of all things that can be a gateway
type Gateway interface {
	// GetGateway returns de DefaultGateway that represents this gateway
	GetGateway() *DefaultGateway

	// SetAttributes sets the free-form attributes
	SetAttributes(attributes map[string]string)

	// SetAntennas sets the antennas
	SetAntennas(antennas []GatewayAntenna)
}

// Value implements sql.Valuer interface
func (r RouterList) Value() (driver.Value, error) {
	routers := make([]string, 0, len(r))
	for _, router := range r {
		routers = append(routers, string(router))
	}
	return strings.Join(routers, ","), nil
}

// Scan implements sql.Scanner interface
func (r *RouterList) Scan(src interface{}) error {
	routers := strings.Split(src.(string), ",")
	result := make(RouterList, 0, len(routers))
	for _, router := range routers {
		result = append(result, router)
	}
	*r = result
	return nil
}

// GetGateway implements Gateway
func (g *DefaultGateway) GetGateway() *DefaultGateway {
	return g
}

// SetAttributes implements Gateway
func (g *DefaultGateway) SetAttributes(attributes map[string]string) {
	g.Attributes = attributes
}

// SetAntennas implements Gateway
func (g *DefaultGateway) SetAntennas(antennas []GatewayAntenna) {
	g.Antennas = antennas
}
