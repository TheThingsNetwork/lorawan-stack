// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import Marshaler from '../util/marshaler'

class ApplicationPackages {
  constructor(registry) {
    this._api = registry
  }

  async getDefaultAssociation(appId, fPort, selector) {
    const result = await this._api.GetDefaultAssociation(
      {
        routeParams: {
          'ids.application_ids.application_id': appId,
          'ids.f_port': fPort,
        },
      },
      Marshaler.selectorToFieldMask(selector),
    )

    return Marshaler.payloadSingleResponse(result)
  }

  async setDefaultAssociation(
    appId,
    fPort,
    patch,
    mask = Marshaler.fieldMaskFromPatch(
      patch,
      this._api.SetDefaultAssociationAllowedFieldMaskPaths,
    ),
  ) {
    const result = await this._api.SetDefaultAssociation(
      {
        routeParams: {
          'default.ids.application_ids.application_id': appId,
          'default.ids.f_port': fPort,
        },
      },
      {
        default: patch,
        field_mask: Marshaler.fieldMask(mask),
      },
    )

    return Marshaler.payloadSingleResponse(result)
  }
}

export default ApplicationPackages
