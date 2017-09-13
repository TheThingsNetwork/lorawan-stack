// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Base represents base component configuration
type Base struct {
	Config []string `name:"config" shorthand:"c" description:"Location of the config files"`
	Log    Log      `name:"log"`
}

// Log represents configuration for the logger
type Log struct {
	Level log.Level `name:"level" description:"The minimum level log messages must have to be shown"`
}

// TLS represents TLS configuration
type TLS struct {
	Certificate string `name:"certificate" description:"Location of TLS certificate"`
	Key         string `name:"key" description:"Location of TLS private key"`
}

// GRPC represents gRPC listener configuration
type GRPC struct {
	Listen    string `name:"listen" description:"Address for the TCP gRPC server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the TLS gRPC server to listen on"`
}

// HTTP represents the HTTP and HTTPS server configuration
type HTTP struct {
	Listen    string `name:"listen" description:"Address for the HTTP server to listen on"`
	ListenTLS string `name:"listen-tls" description:"Address for the HTTPS server to listen on"`
}

// Identity represents identity configuration
type Identity struct {
	Servers map[string]string `name:"servers" description:"TTN Identity Servers (id=https://...)"`
	Keys    map[string]string `name:"keys" description:"TTN Identity Server Public Keys (id=/path/to/...)"`
}

// Redis represents Redis configuration
type Redis struct {
	Address  string `name:"address" description:"Address of the Redis server"`
	Database int    `name:"database" description:"Redis database to use"`
}

// RemoteProviderConfig represents remote config provider configuration(see Viper documentation)
type RemoteProviderConfig struct {
	Name     string `name:"name" description:"Name of the config on the remote without the extension"`
	Provider string `name:"provider" description:"Remote config provider name"`
	Endpoint string `name:"endpoint" description:"Endpoint where the remote config provider is accessible"`
	Path     string `name:"path" description:"Path where to look for the config on the remote"`
	KeyRing  string `name:"keyring" description:"Optional path to secret keyring for initializing a secure connection to remote config provider"`
}

// ServiceBase represents base service configuration
type ServiceBase struct {
	Base         `name:",squash"`
	GRPC         GRPC                  `name:"grpc"`
	HTTP         HTTP                  `name:"http"`
	TLS          TLS                   `name:"tls"`
	Identity     Identity              `name:"identity"`
	RemoteConfig *RemoteProviderConfig `name:"remote-config"`
}

// IsValid returns wether or not the remote config is valid or not
func (c RemoteProviderConfig) IsValid() bool {
	return c.Provider != "" && c.Endpoint != "" && c.Path != ""
}
