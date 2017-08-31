// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// ComponentType represents the type of a component
type ComponentType string

const (
	// Handler is the type of a handler component
	Handler ComponentType = "handler"

	// Router is the type of a router component
	Router ComponentType = "router"

	// Broker is the type of a broker component
	Broker ComponentType = "broker"
)

// DefaultComponent represents a newtork component
type DefaultComponent struct {
	ID      string        `db:"id" json:"id"`
	Type    ComponentType `db:"type" json:"type"`
	Created time.Time     `db:"created" json:"created"`
}

// Component is the interface of all things that can be a network component
type Component interface {
	GetComponent() *DefaultComponent
}

// GetComponent returns the DefaultComponent
func (c *DefaultComponent) GetComponent() *DefaultComponent {
	return c
}

// String implements fmt.Stringer interface
func (t ComponentType) String() string {
	return string(t)
}
