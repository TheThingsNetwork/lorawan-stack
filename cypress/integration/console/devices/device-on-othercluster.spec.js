// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

describe('End device on other cluster', () => {
  const userId = 'cluster-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'cluster-test-user@example.com',
  }

  const applicationId = 'cluster-app-test'
  const application = {
    ids: { application_id: applicationId },
  }

  const deviceId = 'device-all-components'

  const ns = {
    end_device: {
      frequency_plan_id: 'EU_863_870_TTN',
      lorawan_phy_version: 'PHY_V1_0_2_REV_A',
      multicast: false,
      supports_join: true,
      lorawan_version: 'MAC_V1_0_2',
      ids: {
        device_id: deviceId,
        dev_eui: '70B3D57ED8000019',
        join_eui: '0000000000000000',
      },
      supports_class_c: false,
      supports_class_b: false,
      mac_settings: {
        rx2_data_rate_index: 0,
        rx2_frequency: 869525000,
        rx1_delay: 1,
        rx1_data_rate_offset: 0,
        resets_f_cnt: false,
      },
    },
    field_mask: {
      paths: [
        'frequency_plan_id',
        'lorawan_phy_version',
        'multicast',
        'supports_join',
        'lorawan_version',
        'ids.device_id',
        'ids.dev_eui',
        'ids.join_eui',
        'supports_class_c',
        'supports_class_b',
        'mac_settings.rx2_data_rate_index',
        'mac_settings.rx2_frequency',
        'mac_settings.rx1_delay',
        'mac_settings.rx1_data_rate_offset',
        'mac_settings.resets_f_cnt',
      ],
    },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, user.ids.user_id)
    cy.createMockDeviceAllComponents(applicationId, { ns })
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    const response = {
      end_devices: [
        {
          ids: {
            application_ids: {
              application_id: applicationId,
            },
            device_id: deviceId,
          },
          network_server_address: 'tti.staging1.cloud.thethings.industries',
          application_server_address: 'tti.staging1.cloud.thethings.industries',
          join_server_address: 'tti.staging1.cloud.thethings.industries',
        },
      ],
    }

    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/devices`)
    cy.intercept(
      'GET',
      `/api/v3/applications/${applicationId}/devices?page=1&limit=20&field_mask=name,application_server_address,network_server_address,join_server_address`,
      response,
    )
  })

  it('succeeds disabling click on devices that are on another cluster', () => {
    cy.findByText(deviceId).click()
    cy.findByText('End devices (1)')
  })

  it('succeeds showing "Other cluster" status on devices that are on another cluster', () => {
    cy.findByText(deviceId)
      .closest('[role="row"]')
      .within(() => {
        cy.findByText('Other cluster').should('be.visible')
      })
  })

  it('succeeds redirecting when manually accessing devices that are on another cluster', () => {
    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/devices/${deviceId}`,
    )

    const response = {
      application_server_address: 'tti.staging1.cloud.thethings.industries',
      ids: {
        application_ids: {
          application_id: 'cluster-app-test',
        },
        dev_eui: '9000BEEF9000BEEF',
        device_id: deviceId,
        join_eui: '0000000000000000',
      },
      network_server_address: 'tti.staging1.cloud.thethings.industries',
    }

    cy.intercept(
      'GET',
      `api/v3/applications/${applicationId}/devices/${deviceId}?field_mask=name,description,version_ids,network_server_address,application_server_address,join_server_address,locations,attributes`,
      response,
    )

    cy.findByText('Applications (1)')
  })
})
