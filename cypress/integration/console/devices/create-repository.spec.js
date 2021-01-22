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

describe('End device repository create', () => {
  const user = {
    ids: { user_id: 'create-dr-test-user' },
    primary_email_address: 'create-dr-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  describe('OTAA', () => {
    const appId = 'otaa-test-application'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`)
    })

    describe('The Things Products', () => {
      it('succeeds registering the things uno', () => {
        const devId = 'the-things-uno-test'

        // End device selection.
        cy.findByLabelText('Brand').selectOption('the-things-products')
        cy.findByLabelText('Model').selectOption('The Things Uno')
        cy.findByLabelText('Hardware Ver.').selectOption('1.0')
        cy.findByLabelText('Firmware Ver.').selectOption('quickstart')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })

      it('succeeds registering the things node', () => {
        const devId = 'the-things-node-test'

        // End device selection.
        cy.findByLabelText('Brand').selectOption('the-things-products')
        cy.findByLabelText('Model').selectOption('The Things Node')
        cy.findByLabelText('Hardware Ver.').selectOption('1.0')
        cy.findByLabelText('Firmware Ver.').selectOption('1.0')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })

  describe('ABP', () => {
    const appId = 'abp-test-application'
    const application = {
      ids: { application_id: appId },
    }

    before(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`)
    })

    describe('Moko', () => {
      it('succeeds registering lw001-bg device', () => {
        const devId = 'lw001-bg-test'

        // End device selection.
        cy.findByLabelText('Brand').selectOption('moko')
        cy.findByLabelText('Model').selectOption('lw001-bg')
        cy.findByLabelText('Hardware Ver.').selectOption('1.0.1')
        cy.findByLabelText('Firmware Ver.').selectOption('1.0.1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('Device address').type(generateHexValue(8))
        cy.findByLabelText('AppSKey').type(generateHexValue(32))
        cy.findByLabelText('NwkSKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId)

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId}`,
        )
        cy.findByTestId('full-error-view').should('not.exist')
      })
    })
  })
})
