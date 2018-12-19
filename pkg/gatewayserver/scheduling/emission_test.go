// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scheduling_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEmission(t *testing.T) {
	a := assertions.New(t)
	em := scheduling.NewEmission(scheduling.ConcentratorTime(10*time.Second), time.Second)

	a.So(em.Starts(), should.Equal, 10*time.Second)
	a.So(em.Ends(), should.Equal, 11*time.Second)
	a.So(em.Duration(), should.Equal, time.Second)

	a.So(em.Within(scheduling.ConcentratorTime(10*time.Second), scheduling.ConcentratorTime(20*time.Second)), should.Equal, time.Second)
	a.So(em.Within(scheduling.ConcentratorTime(11*time.Second), scheduling.ConcentratorTime(20*time.Second)), should.Equal, 0)
	a.So(em.Within(scheduling.ConcentratorTime(0*time.Second), scheduling.ConcentratorTime(10*time.Second)), should.Equal, time.Second)
	a.So(em.Within(scheduling.ConcentratorTime(0*time.Second), scheduling.ConcentratorTime(11*time.Second)), should.Equal, time.Second)

	a.So(em.OffAir(frequencyplans.TimeOffAir{Fraction: 0.5, Duration: 2 * time.Second}), should.Equal, 2*time.Second)
	a.So(em.OffAir(frequencyplans.TimeOffAir{Fraction: 0.5, Duration: 0}), should.Equal, 500*time.Millisecond)

	toa := frequencyplans.TimeOffAir{Duration: 2 * time.Second}

	// No conflicts.
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(13*time.Second), time.Second), toa), should.Equal, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(13*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(20*time.Second), time.Second), toa), should.Equal, 7*time.Second)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(20*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(7*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(7*time.Second), time.Second), toa), should.Equal, 0)
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(5*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(5*time.Second), time.Second), toa), should.Equal, 2*time.Second)

	// Conflicts.
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(9*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(9*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(10*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(10*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.BeforeWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(11*time.Second), time.Second), toa), should.BeLessThan, 0)
	a.So(em.AfterWithOffAir(scheduling.NewEmission(scheduling.ConcentratorTime(11*time.Second), time.Second), toa), should.BeLessThan, 0)
}

func TestEmissions(t *testing.T) {
	a := assertions.New(t)
	var ems scheduling.Emissions
	ems = ems.Insert(scheduling.NewEmission(scheduling.ConcentratorTime(10*time.Second), time.Second))
	ems = ems.Insert(scheduling.NewEmission(scheduling.ConcentratorTime(6*time.Second), time.Second))
	ems = ems.Insert(scheduling.NewEmission(scheduling.ConcentratorTime(8*time.Second), time.Second))
	ems = ems.Insert(scheduling.NewEmission(scheduling.ConcentratorTime(12*time.Second), time.Second))

	a.So(ems, should.HaveLength, 4)
	a.So(ems[0].Starts(), should.Equal, 6*time.Second)
	a.So(ems[1].Starts(), should.Equal, 8*time.Second)
	a.So(ems[2].Starts(), should.Equal, 10*time.Second)
	a.So(ems[3].Starts(), should.Equal, 12*time.Second)
}
