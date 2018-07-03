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

package rights_test

import (
	"context"
	"errors"
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type myApplicationServer struct{}

func (a *myApplicationServer) MyMethod(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if !ttnpb.IncludesRights(rights.FromContext(ctx), ttnpb.RIGHT_APPLICATION_DEVICES_READ) {
		return nil, errors.New("Not authorized to connect")
	}

	return ttnpb.Empty, nil
}

func ExampleFromContext() {
	s := myApplicationServer{}
	ctx := context.Background()

	_, err := s.MyMethod(ctx, &ttnpb.ApplicationIdentifiers{})
	fmt.Println(err)
	// Output: Not authorized to connect

	ctx = rights.NewContext(ctx, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DEVICES_READ})
	_, err = s.MyMethod(ctx, &ttnpb.ApplicationIdentifiers{})
	fmt.Println(err)
	// <nil>
}
