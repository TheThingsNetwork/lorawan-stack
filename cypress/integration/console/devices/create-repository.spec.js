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

    const selectDevice = ({ brand_id, model_id, hw_version, fw_version, band_id }) => {
      cy.findByLabelText('Brand').selectOption(brand_id)
      cy.findByLabelText('Model').selectOption(model_id)
      cy.findByLabelText('Hardware Ver.').selectOption(hw_version)
      cy.findByLabelText('Firmware Ver.').selectOption(fw_version)
      cy.findByLabelText('Profile (Region)').selectOption(band_id)
    }

    before(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.createApplication(application, user.ids.user_id)
      cy.clearCookies()
      cy.clearLocalStorage()
    })

    beforeEach(() => {
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add/repository`)
    })

    it('displays UI elements in place', () => {
      cy.findByText('Register end device', { selector: 'h1' }).should('be.visible')
      cy.findByRole('button', { name: 'From The LoRaWAN Device Repository' })
        .should('be.visible')
        .and('have.attr', 'href', `/console/applications/${appId}/devices/add/repository`)
      cy.findByRole('button', { name: 'Manually' })
        .should('be.visible')
        .and('have.attr', 'href', `/console/applications/${appId}/devices/add/manual`)

      cy.findByRole('heading', { name: '1. Select the end device' }).should('be.visible')
      cy.findByLabelText('Brand').should('be.visible')
      cy.findByText(/Cannot find your exact end device?/).within(() => {
        cy.findByRole('link', { name: 'Try manual device registration' })
          .should('be.visible')
          .and('have.attr', 'href', `/console/applications/${appId}/devices/add/manual`)
      })

      cy.findByRole('heading', { name: '2. Enter registration data' }).should('be.visible')
      cy.findByText(
        'Please choose an end device first to proceed with entering registration data',
      ).should('be.visible')
      cy.findByRole('button', { name: 'Register end device' })
        .should('be.visible')
        .and('be.disabled')

      cy.findByLabelText('Brand').selectOption('_other_')
      cy.findByTestId('notification')
        .should('be.visible')
        .should('contain', 'Your end device will be added soon!')
        .within(() => {
          cy.findByText(
            /We're sorry, but your device is not yet part of The LoRaWAN Device Repository./,
          ).should('be.visible')
          cy.findByRole('link', { name: 'manual device registration' })
            .should('be.visible')
            .and('have.attr', 'href', `/console/applications/${appId}/devices/add/manual`)
          cy.findByRole('link', { name: /Adding Devices/ })
            .should('be.visible')
            .and('have.attr', 'href', `https://thethingsstack.io/devices/adding-devices/`)
        })

      cy.findByText(
        'Please choose an end device first to proceed with entering registration data',
      ).should('be.visible')
      cy.findByRole('button', { name: 'Register end device' })
        .should('be.visible')
        .and('be.disabled')
    })

    describe('The Things Products', () => {
      const selectUno = () => {
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Uno',
          hw_version: '1.0',
          fw_version: 'quickstart',
          band_id: 'EU_863_870',
        })
      }
      const selectNode = () => {
        selectDevice({
          brand_id: 'the-things-products',
          model_id: 'The Things Node',
          hw_version: '1.0',
          fw_version: '1.0',
          band_id: 'EU_863_870',
        })
      }

      it('displays UI elements in place', () => {
        selectUno()

        cy.findByLabelText('Frequency plan').should('be.visible')
        cy.findDescriptionByLabelText('Frequency plan')
          .should('contain', 'The frequency plan used by the end device')
          .and('be.visible')
        cy.findByLabelText('DevEUI').should('be.visible')
        cy.findDescriptionByLabelText('DevEUI')
          .should('contain', 'The DevEUI is the unique identifier for this end device')
          .and('be.visible')
        cy.findByLabelText('AppEUI').should('be.visible')
        cy.findDescriptionByLabelText('AppEUI')
          .should(
            'contain',
            'The AppEUI uniquely identifies the owner of the end device. If no AppEUI is provided by the device manufacturer (usually for development), it can be filled with zeros.',
          )
          .and('be.visible')
        cy.findByLabelText('AppKey').should('be.visible')
        cy.findDescriptionByLabelText('AppKey')
          .should(
            'contain',
            'The root key to derive session keys to secure communication between the end device and the application',
          )
          .and('be.visible')
        cy.findByLabelText('End device ID').should('be.visible')
        cy.findByLabelText('View registered end device').should('be.visible')
        cy.findByLabelText('Register another end device of this type').should('be.visible')

        cy.findByRole('button', { name: 'Register end device' })
          .should('be.visible')
          .and('not.be.disabled')
      })

      it('succeeds registering the things uno', () => {
        const devId = 'the-things-uno-test'

        selectUno()

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

        selectNode()

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

      it('succeeds registering two the things uno', () => {
        const devId1 = 'the-things-uno-test-1'

        // End device selection.
        selectUno()

        // End device registration.
        cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
        cy.findByLabelText('AppEUI').type(generateHexValue(16))
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId1)
        cy.findByLabelText('Register another end device of this type').check()

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.findByTestId('toast-notification')
          .should('be.visible')
          .findByText('End device registered')
          .should('be.visible')

        const devId2 = 'the-things-uno-test-2'

        // End device registration.
        cy.findByLabelText('DevEUI').type(generateHexValue(16))
        cy.findByLabelText('AppKey').type(generateHexValue(32))
        cy.findByLabelText('End device ID').type(devId2)
        cy.findByLabelText('View registered end device').check()

        cy.findByRole('button', { name: 'Register end device' }).click()

        cy.location('pathname').should(
          'eq',
          `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${devId2}`,
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
      it('displays UI elements in place', () => {
        cy.findByLabelText('Brand').selectOption('moko')
        cy.findByLabelText('Model').selectOption('lw001-bg')
        cy.findByLabelText('Hardware Ver.').selectOption('1.0.1')
        cy.findByLabelText('Firmware Ver.').selectOption('1.0.1')
        cy.findByLabelText('Profile (Region)').selectOption('EU_863_870')

        cy.findByLabelText('Frequency plan').should('be.visible')
        cy.findDescriptionByLabelText('Frequency plan')
          .should('contain', 'The frequency plan used by the end device')
          .and('be.visible')
        cy.findByLabelText('Device address').should('be.visible')
        cy.findDescriptionByLabelText('Device address')
          .should(
            'contain',
            'Device address, issued by the Network Server or chosen by device manufacturer in case of testing range',
          )
          .and('be.visible')
        cy.findByLabelText('AppSKey').should('be.visible')
        cy.findDescriptionByLabelText('AppSKey')
          .should('contain', 'Application session key')
          .and('be.visible')
        cy.findByLabelText('NwkSKey').should('be.visible')
        cy.findDescriptionByLabelText('NwkSKey')
          .should('contain', 'Network session key')
          .and('be.visible')
        cy.findByLabelText('End device ID').should('be.visible')
        cy.findByLabelText('View registered end device').should('be.visible')
        cy.findByLabelText('Register another end device of this type').should('be.visible')

        cy.findByRole('button', { name: 'Register end device' })
          .should('be.visible')
          .and('not.be.disabled')
      })

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
