// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package frequencyplan contains the structs to handle frequency plans
package frequencyplan

import (
	"github.com/TheThingsNetwork/ttn/pkg/band"
)

// Band is a type covering band.Band, and that translates in yaml to a string with the ID of a band. Using this type in the frequency plan avoids having to marshal the whole regional band.
type Band band.Band

// UnmarshalYAML implements gopkg.in/yaml.v2.Unmarshaler
func (b *Band) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var strBand string
	if err := unmarshal(&strBand); err != nil {
		return err
	}

	newBand, err := band.GetByID(band.BandID(strBand))
	if err != nil {
		return err
	}

	*b = Band(newBand)
	return nil
}

// MarshalYAML implements gopkg.in/yaml.v2.Marshaler
func (b Band) MarshalYAML() (interface{}, error) {
	return b.ID, nil
}

// FrequencyPlan contains the frequency plan
type FrequencyPlan struct {
	Band     Band       `yaml:"band"`
	Channels []Channel  `yaml:"channels"`
	LBT      *LBTConfig `yaml:"lbt,omitempty"`
	Radios   []Radio    `yaml:"radios"`
}

// Channel describes one of the channels of the frequency plan
type Channel struct {
	Frequency uint32 `yaml:"frequency"`
	DataRate  *uint8 `yaml:"datarate,omitempty"`
}

// LBTConfig describes listen-before-talk configuration if applicable in the frequency plan
type LBTConfig struct {
	RSSIOffset float32  `yaml:"rssi_offset,omitempty"`
	RSSITarget *float32 `yaml:"rssi_target,omitempty"`
	ScanTime   *int32   `yaml:"scan_time,omitempty"`
}

// Radio describes the configuration of a radio
type Radio struct {
	Frequency uint32 `yaml:"frequency"`
	// TX is nil if the radio has disabled TX emission
	TX *RadioTXConfig `yaml:"tx,omitempty"`
}

// RadioTXConfig describes the TX configuration for a radio that has enabled TX emission
type RadioTXConfig struct {
	MinFrequency   uint32  `yaml:"min_frequency"`
	MaxFrequency   uint32  `yaml:"max_frequency"`
	NotchFrequency *uint32 `yaml:"notch_frequency,omitempty"`
}
