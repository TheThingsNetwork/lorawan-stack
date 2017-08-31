// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
)

// DefaultApplication represents an application on the network
type DefaultApplication struct {
	// ID is the unique id of the application
	ID string `json:"id" db:"id"`

	// Description is the description of the application
	Description string `json:"name" db:"description"`

	// EUIs are the app euis this application uses
	EUIs []AppEUI `json:"euis"`

	// APIKeys are the API keys the application defined
	APIKeys []ApplicationAPIKey `json:"api_keys"`

	// Created is the time when the user account was created
	Created time.Time `json:"created" db:"created"`

	// Deleted is the time when the user account was deleted
	Archived *time.Time `json:"deleted,omitempty" db:"archived"`
}

// AppEUI is a type that overloads the base EUI64 type to implemen the sql.Scanner
// and sql.Valuer interface
type AppEUI types.EUI64

// ApplicationAPIKey represents an API key of an application
type ApplicationAPIKey struct {
	// Name is the API key name
	Name string `db:"name" json:"name"`

	// Key is the actual API key (base64 encoded)
	Key string `db:"key" json:"key"`

	// Rights are the rights this API key bears
	Rights []Right `json:"rights"`
}

// Application is the interface all things that can be an application
// This can be used to build richer user types that can still be
// read and written to a database.
type Application interface {
	// GetApplication returns the DefaultApplication that represents this
	// application
	GetApplication() *DefaultApplication

	// SetEUIs sets the apps EUIs
	SetEUIs([]AppEUI)

	// SetAPIKeys sets the apps APIKeys
	SetAPIKeys([]ApplicationAPIKey)
}

// GetApplication implements Application
func (d *DefaultApplication) GetApplication() *DefaultApplication {
	return d
}

// SetEUIs implements Application
func (d *DefaultApplication) SetEUIs(euis []AppEUI) {
	d.EUIs = euis
}

// SetAPIKeys implements Application
func (d *DefaultApplication) SetAPIKeys(apiKeys []ApplicationAPIKey) {
	d.APIKeys = apiKeys
}

// Value implements sql.Valuer interface
func (e AppEUI) Value() (driver.Value, error) {
	eui := types.EUI64(e)
	return eui.MarshalText()
}

// Scan implements sql.Scanner interface
func (e *AppEUI) Scan(src interface{}) error {
	eui := new(types.EUI64)
	err := eui.UnmarshalText(src.([]byte))
	if err != nil {
		return err
	}
	*e = AppEUI(*eui)
	return nil
}
