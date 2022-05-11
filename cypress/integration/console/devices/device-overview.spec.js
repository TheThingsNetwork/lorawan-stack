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

describe('Device overview', () => {
  const appId = 'end-device-mac-data-test-application'
  const application = { ids: { application_id: appId } }
  const userId = 'end-device-mac-data-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'end-device-mac-data-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const ns = {
    end_device: {
      frequency_plan_id: 'EU_863_870_TTN',
      lorawan_phy_version: 'PHY_V1_0_2_REV_B',
      multicast: false,
      supports_join: false,
      lorawan_version: 'MAC_V1_0_2',
      ids: {
        device_id: 'device-all-components',
        dev_eui: '70B3D57ED8000019',
      },
      session: {
        keys: {
          f_nwk_s_int_key: {
            key: 'CBFBF585D81A9063A31EA6922EDD6360',
          },
        },
        dev_addr: '270000FC',
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
  }
  const endDeviceId = ns.end_device.ids.device_id

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createMockDeviceAllComponents(appId, undefined, { ns })
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${endDeviceId}`)
  })

  it('succeeds downloading device MAC state', () => {
    cy.findByRole('button', { name: /Download MAC data/ }).click()
    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Download MAC data', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Download MAC data/ }).click()
      })
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').should('not.exist')
  })

  it('succeeds showing warning when there is no MAC state', () => {
    const response = {
      ids: {
        device_id: endDeviceId,
        application_ids: {
          application_id: appId,
        },
        dev_eui: '70B3D57ED8000019',
        dev_addr: '270000FC',
      },
      created_at: '2021-07-06T21:32:48.499001538Z',
      updated_at: '2022-05-04T10:54:11.944463141Z',
    }

    cy.intercept(
      'GET',
      `/api/v3/ns/applications/${appId}/devices/${endDeviceId}?field_mask=mac_state`,
      response,
    )

    cy.findByRole('button', { name: /Download MAC data/ }).click()
    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Download MAC data', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Download MAC data/ }).click()
      })
    cy.findByTestId('toast-notification')
      .findByText(`There was an error and MAC state could not be included in the MAC data.`)
      .should('be.visible')
  })
})
