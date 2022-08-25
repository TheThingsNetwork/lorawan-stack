// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package profilefetcher_test

import (
	"context"

	mockdr "go.thethings.network/lorawan-stack/v3/pkg/devicerepository/mock"
	. "go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter/profilefetcher"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type mockTemplateFetcher struct {
	mockDR *mockdr.MockDR
}

// GetTemplate makes a request to the Device Repository server with its predefined call options.
func (tf *mockTemplateFetcher) GetTemplate(
	ctx context.Context,
	in *ttnpb.GetTemplateRequest,
) (*ttnpb.EndDeviceTemplate, error) {
	return tf.mockDR.GetTemplate(ctx, in)
}

// MockTemplateFetcher returns an end-device template fetcher that directly points to the DR mock.
func MockTemplateFetcher(mockDR *mockdr.MockDR) TemplateFetcher {
	return &mockTemplateFetcher{
		mockDR: mockDR,
	}
}
