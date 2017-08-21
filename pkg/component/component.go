// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package component

import (
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Config is the type of configuration for Components
type Config struct {
	config.ServiceBase `name:",squash"`
}

// Component is a base component for The Things Network cluster
type Component struct {
	config *Config
	log    log.Interface
}

// New returns a new component
func New(log log.Interface, config *Config) *Component {
	return &Component{
		config: config,
		log:    log.WithField("component", "base"),
	}
}

// Start starts the component
func (c *Component) Start() {
	c.log.Debug("Starting component")
}
