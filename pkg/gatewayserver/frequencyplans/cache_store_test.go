// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans_test

import (
	"errors"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/frequencyplans"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var (
	validFrequencyPlanID   = "validID"
	invalidFrequencyPlanID = "invalidID"

	allIDs = []string{validFrequencyPlanID, invalidFrequencyPlanID}
)

type DummyStore struct {
	hitIDs chan bool
	hitFP  chan bool
}

func (d DummyStore) GetAllIDs() []string {
	select {
	case d.hitIDs <- true:
	default:
	}
	return allIDs
}

func (d DummyStore) GetByID(id string) (ttnpb.FrequencyPlan, error) {
	if id == validFrequencyPlanID {
		select {
		case d.hitFP <- true:
		default:
		}

		return ttnpb.FrequencyPlan{}, nil
	}

	return ttnpb.FrequencyPlan{}, errors.New("Invalid frequency plan ID")
}

func TestCacheStore(t *testing.T) {
	dummy := DummyStore{
		hitIDs: make(chan bool),
		hitFP:  make(chan bool),
	}
	store := frequencyplans.Cache(dummy, frequencyplans.DefaultCacheExpiry)

	stored := make(chan bool)

	var ids []string
	go func() {
		ids = store.GetAllIDs()
		stored <- true
	}()

	select {
	case <-dummy.hitIDs:
		<-stored
	case <-stored:
		t.Log("Cache did not hit on the original store to retrieve the list of frequency plan IDs, even though there should not be any stored value yet")
		t.Fail()
	}

	go func() {
		store.GetAllIDs()
		stored <- true
	}()

	select {
	case <-dummy.hitIDs:
		t.Log("Cache hit on the original store to retrieve the list of frequency plan IDs, whereas it should have used cached value")
		t.Fail()
	case <-stored:
	}

	for _, id := range ids {
		if id != validFrequencyPlanID && id != invalidFrequencyPlanID {
			t.Log("Unknown frequency plan ID returned by cache")
			t.Fail()
		}
	}

	go func() {
		store.GetByID(validFrequencyPlanID)
		stored <- true
	}()

	select {
	case <-dummy.hitFP:
		<-stored
	case <-stored:
		t.Log("Cache did not hit on the original store to retrieve the frequency plan, even though there should not be any stored value yet")
		t.Fail()
	}

	go func() {
		store.GetByID(validFrequencyPlanID)
		stored <- true
	}()

	select {
	case <-dummy.hitFP:
		t.Log("Cache hit on the original store to retrieve the frequency plan, whereas it should have used cached value")
		t.Fail()
	case <-stored:
	}
}

func BenchmarkCacheGetByID(b *testing.B) {
	dummy := DummyStore{}
	store := frequencyplans.Cache(dummy, frequencyplans.DefaultCacheExpiry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.GetByID(validFrequencyPlanID); err != nil {
			b.Error("Unexpected return:", err)
		}
	}
}

func BenchmarkCacheGetAllIDs(b *testing.B) {
	dummy := DummyStore{}
	store := frequencyplans.Cache(dummy, frequencyplans.DefaultCacheExpiry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetAllIDs()
	}
}
