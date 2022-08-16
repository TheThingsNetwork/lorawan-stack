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

import { generateHexValue } from '../../../support/utils'

describe('Device general settings', () => {
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
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/devices/${endDeviceId}/general-settings`,
    )
  })

  it('displays newly created end device values', () => {
    cy.findByRole('heading', { name: 'Basic' }).should('be.visible')
    cy.findByLabelText('End device ID')
      .should('be.disabled')
      .and('have.attr', 'value')
      .and('eq', endDeviceId)
    cy.findByLabelText(/AppEUI/).should('be.disabled')
    cy.findByLabelText('DevEUI')
      .should('be.disabled')
      .and('have.attr', 'value')
      .and('eq', ns.end_device.ids.dev_eui)

    cy.fixture('console/devices/device.is.json').then(endDevice => {
      cy.findByLabelText('End device name')
        .should('be.visible')
        .and('have.attr', 'value', endDevice.end_device.name)
      cy.findByLabelText('End device description')
        .should('be.visible')
        .and('have.text', endDevice.end_device.description)
      cy.findDescriptionByLabelText('End device description')
        .should(
          'contain',
          'Optional end device description; can also be used to save notes about the end device',
        )
        .and('be.visible')
      cy.findByLabelText('Network Server address')
        .should('be.visible')
        .and('have.attr', 'value', window.location.hostname)
      cy.findByLabelText('Application Server address')
        .should('be.visible')
        .and('have.attr', 'value', window.location.hostname)
    })
  })

  it('succeeds editing device name and description', () => {
    cy.findByLabelText(/name/).type(' Updated')
    cy.findByLabelText(/description/).type('Updated ')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`End device updated`).should('be.visible')
  })

  it('succeeds editing Network layer', () => {
    const device = {
      dev_addr: generateHexValue(8),
      nwk_s_key: generateHexValue(32),
      lorawan_version: 'MAC_V1_0',
      frequency_plan_id: '863-870 MHz',
    }

    cy.findByText('Network layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('Frequency plan').type(device.frequency_plan_id)
        cy.findByLabelText('LoRaWAN version').type(device.lorawan_version)
        cy.findByLabelText('Device address').type(device.dev_addr)
        cy.findByLabelText('NwkSKey').type(device.nwk_s_key)
        cy.findByRole('button', { name: 'Save changes' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`End device updated`).should('be.visible')
  })

  it('succeeds editing Application layer', () => {
    const device = {
      app_s_key: generateHexValue(32),
    }

    cy.findByText('Application layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('AppSKey').type(device.app_s_key)
        cy.findByRole('button', { name: 'Save changes' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`End device updated`)
      .should('be.visible')
  })

  it('succeeds editing Join Settings', () => {
    const device = {
      home_netid: generateHexValue(6),
      server_id: 'test-server-id',
      app_key: generateHexValue(32),
    }

    cy.findByText('Join settings', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('Home NetID').type(device.home_netid)
        cy.findByLabelText('Application Server ID').type(device.server_id)
        cy.findByLabelText('AppKey').type(device.app_key)
        cy.findByRole('button', { name: 'Save changes' }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`End device updated`).should('be.visible')
  })

  it('succeeds editing server adresses', () => {
    cy.findByLabelText('Network Server address').type('.test')
    cy.findByLabelText('Application Server address').type('.test')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`End device updated`).should('be.visible')
  })

  it('succeeds adding end device attributes', () => {
    cy.findByRole('button', { name: /Add attributes/ }).click()

    cy.get(`[name="attributes[0].key"]`).type('end-device-test-key')
    cy.get(`[name="attributes[0].value"]`).type('end-device-test-value')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`End device updated`).should('be.visible')
  })

  it('succeeds deleting end device', () => {
    cy.findByRole('button', { name: /Delete end device/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Delete end device', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: /Delete end device/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/devices`,
    )

    cy.findByRole('cell', { name: appId }).should('not.exist')
  })
})
