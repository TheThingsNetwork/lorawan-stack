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

import { defineSmokeTest } from '../utils'

const checkCollapsingFields = defineSmokeTest('check all end device sub pages', () => {
  const userId = 'device-subpage-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'device-subpage-test-user@example.com',
  }

  const applicationId = 'collapsing-fields-app-test'
  const application = {
    ids: { application_id: applicationId },
  }
  const deviceId = 'device-all-components'
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

  cy.createUser(user)
  cy.createApplication(application, user.ids.user_id)
  cy.createMockDeviceAllComponents(applicationId, undefined, { ns })
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Applications/ }).click()
  })
  cy.findByRole('cell', { name: application.ids.application_id }).click()
  cy.findByRole('cell', { name: deviceId }).click()

  cy.get('#stage').within(() => {
    cy.findAllByText(deviceId).should('be.visible')
    cy.findByRole('columnheader', { name: 'General information' }).parent().should('be.visible')
    cy.findByRole('columnheader', { name: 'Activation information' }).parent().should('be.visible')
    cy.findByRole('columnheader', { name: 'Session information' }).parent().should('be.visible')
    cy.findByRole('columnheader', { name: 'MAC data' }).parent().should('be.visible')
    cy.findByRole('heading', { name: /Live data/ })
      .parent()
      .should('be.visible')
    cy.findByRole('heading', { name: /Location/ })
      .parent()
      .should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByRole('button', { name: 'Live data' }).click()
    cy.findByText(/Waiting for events from/).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByRole('button', { name: 'Messaging' }).click()
    cy.findByRole('heading', { name: 'Simulate uplink' }).should('be.visible')
    cy.findByRole('button', { name: 'Simulate uplink' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByRole('button', { name: 'Downlink' }).click()
    cy.findByRole('heading', { name: 'Schedule downlink' }).should('be.visible')
    cy.findByRole('button', { name: 'Schedule downlink' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByRole('button', { name: 'Location' }).click()
    cy.findByRole('heading', { name: 'Set end device location manually' }).should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByRole('button', { name: 'Payload formatters' }).click()
    cy.findByRole('button', { name: /Uplink/ }).should('be.visible')
    cy.findByLabelText('Formatter type').should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByRole('button', { name: /Downlink/ }).click()
    cy.findByLabelText('Formatter type').should('be.visible')
    cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
    cy.findByTestId('error-notification').should('not.exist')

    cy.findByRole('button', { name: /General settings/ }).click()
    cy.findByText('Basic')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByLabelText('End device ID')
        cy.findByRole('button', { name: /Save changes/ }).should('be.visible')
        cy.findByTestId('error-notification').should('not.exist')
        cy.findByRole('button', { name: 'Collapse' }).click()
      })
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('Frequency plan').should('be.visible')
        cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
        cy.findByText('Advanced MAC settings').click()
        cy.findByText('Frame counter width').should('be.visible')
        cy.findByTestId('error-notification').should('not.exist')
        cy.findByRole('button', { name: 'Collapse' }).click()
      })
    cy.findByText('Application layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText('Payload crypto override').should('be.visible')
        cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
        cy.findByTestId('error-notification').should('not.exist')
        cy.findByRole('button', { name: 'Collapse' }).click()
      })
    cy.findByText('Join settings')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByLabelText('Home NetID').should('be.visible')
        cy.findByRole('button', { name: 'Save changes' }).should('be.visible')
        cy.findByTestId('error-notification').should('not.exist')
        cy.findByRole('button', { name: 'Collapse' }).click()
      })
    cy.findByTestId('error-notification').should('not.exist')
  })
})

export default [checkCollapsingFields]
