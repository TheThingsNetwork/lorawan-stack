// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/store"
)

func filterPrefixBytes(m map[string][]byte, prefix string) map[string][]byte {
	prefix += store.Separator
	ret := make(map[string][]byte, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			ret[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return ret
}

func filterPrefix(m map[string]interface{}, prefix string) map[string]interface{} {
	prefix += store.Separator
	ret := make(map[string]interface{}, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			ret[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return ret
}

func (m *Message) UnmarshalMap(im map[string]interface{}) error {
	if err := store.UnmarshalMap(filterPrefix(im, "MHDR"), &m.MHDR); err != nil {
		return err
	}
	if err := store.UnmarshalMap(filterPrefix(im, "MIC"), &m.MIC); err != nil {
		return err
	}
	var pld isMessage_Payload
	switch m.MHDR.MType {
	case MType_UNCONFIRMED_DOWN,
		MType_UNCONFIRMED_UP,
		MType_CONFIRMED_UP,
		MType_CONFIRMED_DOWN:
		pld = &Message_MACPayload{}
	case MType_JOIN_REQUEST:
		pld = &Message_JoinRequestPayload{}
	case MType_JOIN_ACCEPT:
		pld = &Message_JoinAcceptPayload{}
	}
	return store.UnmarshalMap(filterPrefix(im, "Payload"), pld)
}

func (m *Message) UnmarshalByteMap(bm map[string][]byte) error {
	if err := store.UnmarshalByteMap(filterPrefixBytes(bm, "MHDR"), &m.MHDR); err != nil {
		return err
	}
	if err := store.UnmarshalByteMap(filterPrefixBytes(bm, "MIC"), &m.MIC); err != nil {
		return err
	}
	var pld isMessage_Payload
	switch m.MHDR.MType {
	case MType_UNCONFIRMED_DOWN,
		MType_UNCONFIRMED_UP,
		MType_CONFIRMED_UP,
		MType_CONFIRMED_DOWN:
		pld = &Message_MACPayload{}
	case MType_JOIN_REQUEST:
		pld = &Message_JoinRequestPayload{}
	case MType_JOIN_ACCEPT:
		pld = &Message_JoinAcceptPayload{}
	}
	return store.UnmarshalByteMap(filterPrefixBytes(bm, "Payload"), pld)
}
