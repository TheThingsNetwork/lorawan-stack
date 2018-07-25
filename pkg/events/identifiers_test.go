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

package events_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestIdentifierFilter(t *testing.T) {
	a := assertions.New(t)

	ch := make(events.Channel, 10)

	filter := events.NewIdentifierFilter()

	evtAppFoo := events.New(test.Context(), "test", ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}, "hello foo")
	evtAppBar := events.New(test.Context(), "test", ttnpb.ApplicationIdentifiers{ApplicationID: "bar"}, "hello bar")

	evtCliFoo := events.New(test.Context(), "test", ttnpb.ClientIdentifiers{ClientID: "foo"}, "hello foo")
	evtCliBar := events.New(test.Context(), "test", ttnpb.ClientIdentifiers{ClientID: "bar"}, "hello bar")

	evtDevFoo := events.New(test.Context(), "test", ttnpb.EndDeviceIdentifiers{
		DeviceID:               "foo",
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
	}, "hello foo")
	evtDevBar := events.New(test.Context(), "test", ttnpb.EndDeviceIdentifiers{
		DeviceID:               "bar",
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "bar"},
	}, "hello bar")

	evtGtwFoo := events.New(test.Context(), "test", ttnpb.GatewayIdentifiers{GatewayID: "foo"}, "hello foo")
	evtGtwBar := events.New(test.Context(), "test", ttnpb.GatewayIdentifiers{GatewayID: "bar"}, "hello bar")

	evtOrgFoo := events.New(test.Context(), "test", ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}, "hello foo")
	evtOrgBar := events.New(test.Context(), "test", ttnpb.OrganizationIdentifiers{OrganizationID: "bar"}, "hello bar")

	evtUsrFoo := events.New(test.Context(), "test", ttnpb.UserIdentifiers{UserID: "foo"}, "hello foo")
	evtUsrBar := events.New(test.Context(), "test", ttnpb.UserIdentifiers{UserID: "bar"}, "hello bar")

	fooIDs := &ttnpb.CombinedIdentifiers{
		ApplicationIDs:  []*ttnpb.ApplicationIdentifiers{{ApplicationID: "foo"}},
		ClientIDs:       []*ttnpb.ClientIdentifiers{{ClientID: "foo"}},
		DeviceIDs:       []*ttnpb.EndDeviceIdentifiers{{DeviceID: "foo", ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"}}},
		GatewayIDs:      []*ttnpb.GatewayIdentifiers{{GatewayID: "foo"}},
		OrganizationIDs: []*ttnpb.OrganizationIdentifiers{{OrganizationID: "foo"}},
		UserIDs:         []*ttnpb.UserIdentifiers{{UserID: "foo"}},
	}

	filter.Subscribe(test.Context(), fooIDs, ch)

	filter.Notify(evtAppBar)
	filter.Notify(evtAppFoo)

	a.So(<-ch, should.Equal, evtAppFoo)
	a.So(ch, should.BeEmpty)

	filter.Notify(evtCliBar)
	filter.Notify(evtCliFoo)

	a.So(<-ch, should.Equal, evtCliFoo)
	a.So(ch, should.BeEmpty)

	filter.Notify(evtDevBar)
	filter.Notify(evtDevFoo)

	a.So(<-ch, should.Equal, evtDevFoo)
	a.So(ch, should.BeEmpty)

	filter.Notify(evtGtwBar)
	filter.Notify(evtGtwFoo)

	a.So(<-ch, should.Equal, evtGtwFoo)
	a.So(ch, should.BeEmpty)

	filter.Notify(evtOrgBar)
	filter.Notify(evtOrgFoo)

	a.So(<-ch, should.Equal, evtOrgFoo)
	a.So(ch, should.BeEmpty)

	filter.Notify(evtUsrBar)
	filter.Notify(evtUsrFoo)

	a.So(<-ch, should.Equal, evtUsrFoo)
	a.So(ch, should.BeEmpty)

	filter.Unsubscribe(test.Context(), fooIDs, ch)

	filter.Notify(evtAppFoo)
	filter.Notify(evtCliFoo)
	filter.Notify(evtDevFoo)
	filter.Notify(evtGtwFoo)
	filter.Notify(evtOrgFoo)
	filter.Notify(evtUsrFoo)

	if !a.So(ch, should.BeEmpty) {
	loop:
		for {
			select {
			case evt := <-ch:
				fmt.Println(evt)
			default:
				break loop
			}
		}
	}
}
