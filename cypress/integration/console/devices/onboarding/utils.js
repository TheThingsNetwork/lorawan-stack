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

export const selectDevice = ({ brand_id, model_id, hw_version, fw_version, band_id }) => {
  cy.findByLabelText('End device brand').selectOption(brand_id)
  cy.findByLabelText('Model').selectOption(model_id)
  cy.findByLabelText('Hardware Ver.').selectOption(hw_version)
  cy.findByLabelText('Firmware Ver.').selectOption(fw_version)
  cy.findByLabelText('Profile (Region)').selectOption(band_id)
}

export const interceptDeviceRepo = appId => {
  cy.intercept(
    'GET',
    `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models/test-model4/1.0/EU_863_870/template`,
    { fixture: 'console/devices/repository/test-brand-otaa-model4.template.json' },
  )
  cy.intercept(
    'GET',
    `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models/test-model3/1.0.1/EU_863_870/template`,
    { fixture: 'console/devices/repository/test-brand-otaa-model3.template.json' },
  )
  cy.intercept(
    'GET',
    `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models/test-model2/1.0/EU_863_870/template`,
    { fixture: 'console/devices/repository/test-brand-otaa-model2.template.json' },
  )
  cy.intercept('GET', `/api/v3/dr/applications/${appId}/brands/test-brand-otaa/models*`, {
    fixture: 'console/devices/repository/test-brand-otaa.models.json',
  })
  cy.intercept('GET', `/api/v3/dr/applications/${appId}/brands*`, {
    fixture: 'console/devices/repository/brands.json',
  })
}

export const composeClaimResponse = ({ joinEui, devEui, id, appId }) => ({
  application_ids: { application_id: appId },
  device_id: id,
  dev_eui: devEui,
  join_eui: joinEui,
  dev_addr: '2600ABCD',
})

export const composeExpectedRequest = ({ joinEui, devEui, cac, id, appId }) => ({
  authenticated_identifiers: {
    join_eui: joinEui.toUpperCase(),
    dev_eui: devEui.toUpperCase(),
    authentication_code: cac,
  },
  target_device_id: id,
  target_application_ids: { application_id: appId },
})
