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

const checkCollapsingFields = defineSmokeTest(
  'succeeds showing contents of collapsing fields in device general settings',
  () => {
    const userId = 'collapsing-fields-test-user'
    const user = {
      ids: { user_id: userId },
      primary_email_address: 'collapsing-fields-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }

    const applicationId = 'collapsing-fields-app-test'
    const application = {
      ids: { application_id: applicationId },
      name: 'Application End Devices Test Name',
      description: 'Application End Devices Test Description',
    }

    const endDeviceId = 'collapsing-fields-end-device-test'
    const endDevice = {
      application_server_address: 'localhost',
      ids: {
        device_id: endDeviceId,
        dev_eui: '0000000000000001',
        join_eui: '0000000000000000',
      },
      name: 'End Device Test Name',
      description: 'End Device Test Description',
      join_server_address: 'localhost',
      network_server_address: 'localhost',
    }
    const endDeviceFieldMask = {
      paths: [
        'join_server_address',
        'network_server_address',
        'application_server_address',
        'ids.dev_eui',
        'ids.join_eui',
        'name',
        'description',
      ],
    }
    const endDeviceRequestBody = {
      end_device: endDevice,
      field_mask: endDeviceFieldMask,
    }

    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createEndDevice(applicationId, endDeviceRequestBody)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${applicationId}/devices/${endDeviceId}/general-settings`,
    )
    cy.findByText('Network layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' })
          .invoke('attr', 'disabled')
          .then(disabled => {
            if (disabled) {
              cy.log('buttonIsDiabled')
            } else {
              cy.findByRole('button', { name: 'Expand' }).click()
              cy.findByText('Frequency plan')
              cy.findByText('Advanced MAC settings').click()
              cy.findByText('Frame counter width')
            }
          })
      })

    cy.findByText('Application layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' })
          .invoke('attr', 'disabled')
          .then(disabled => {
            if (disabled) {
              cy.log('buttonIsDiabled')
            } else {
              cy.findByRole('button', { name: 'Expand' }).click()
              cy.findByText('Payload crypto override')
            }
          })
      })

    cy.findByText('Join Settings', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' })
          .invoke('attr', 'disabled')
          .then(disabled => {
            if (disabled) {
              cy.log('buttonIsDiabled')
            } else {
              cy.findByRole('button', { name: 'Expand' }).click()
              cy.findByText('Home NetID')
            }
          })
      })
  },
)

export default [checkCollapsingFields]
