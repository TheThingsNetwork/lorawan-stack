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

describe('Device un-claiming', () => {
  const appId = 'end-device-edit-test-application'
  const application = { ids: { application_id: appId } }
  const userId = 'end-device-edit-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'end-device-edit-test-user@example.com',
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
    cy.intercept('POST', '/api/v3/edcs/claim/info', { body: { supports_claiming: true } })

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/devices/${endDeviceId}/general-settings`,
    )
  })

  it('succeeds un-claiming and deleting an end device', () => {
    cy.intercept('DELETE', `/api/v3/edcs/claim/${appId}/devices/${endDeviceId}?*`, {}).as(
      'unclaim-request',
    )

    cy.findByRole('button', { name: /Unclaim and delete end device/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Unclaim and delete end device', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Unclaim and delete end device/ }).click()
      })

    cy.wait('@unclaim-request')
      .its('request.url')
      .then(url => {
        const params = new URLSearchParams(new URL(url).search)
        const hexToBase64 = hex =>
          btoa(String.fromCharCode(...hex.match(/.{1,2}/g).map(byte => parseInt(byte, 16))))

        expect(params.get('dev_eui')).to.equal(hexToBase64(ns.end_device.ids.dev_eui))
        expect(params.get('join_eui')).to.equal(hexToBase64('0000000000000000'))
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/devices`,
    )

    cy.findByRole('cell', { name: appId }).should('not.exist')
  })
})
