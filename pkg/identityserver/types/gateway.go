// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"strings"
	"time"
)

// RouterList is the type that represent a list of routers a gateway connect to
type RouterList []string

// Placement is the type of antenna placement
type Placement string

const (
	// Indoor means that the antenna is placed indoors
	Indoor Placement = "indoor"

	// Outdoor means that the antenna is placed outdoors
	Outdoor Placement = "outdoor"
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

	// StatusPublic denotes wether or not the gateway's status is public or not
	StatusPublic bool `db:"status_public"`

	// LocationPublic denotes wether or not the gateway's location is public
	LocationPublic bool `db:"location_public"`

	// OwnerPublic denotes wether or not the gateway owner is public
	OwnerPublic bool `db:"owner_public"`

	// AutoUpdate indicates wether or not the gateway should be able to
	// automatically fetch and execute firmware updates
	AutoUpdate bool `db:"auto_update"`

	// Brand is the gateway brand
	Brand *string `db:"brand"`

	// Model is the gateway model
	Model *string `db:"model"`

	// AntennaType denotes the antenna's type
	AntennaType *string `db:"antenna_type"`

	// AntennaModel denotes the antenna's model
	AntennaModel *string `db:"antenna_model"`

	// AntennaPlacement denotes wether if antenna is placed indoors or outdoors
	AntennaPlacement *Placement `db:"antenna_placement"`

	// AntennaAltitude denotes the estimated height the antenna is placed
	AntennaAltitude *string `db:"antenna_altitude"`

	// AntennaLocation denotes the
	AntennaLocation *string `db:"antenna_location"`

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

// Gateway is the interface of all things that can be a gateway
type Gateway interface {
	// GetGateway returns de DefaultGateway that represents this gateway
	GetGateway() *DefaultGateway

	// SetAttributes sets the free-form attributes
	SetAttributes(attributes map[string]string)
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
