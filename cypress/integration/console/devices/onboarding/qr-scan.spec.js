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

import { generateHexValue } from '../../../../support/utils'

import {
  interceptDeviceRepo,
  selectDevice,
  composeClaimResponse,
  composeExpectedRequest,
} from './utils'

describe('Device onboarding with QR scan', () => {
  const user = {
    ids: { user_id: 'create-manual-test-user' },
    primary_email_address: 'create-manual-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const appId = 'qr-test-application'
  const application = {
    ids: { application_id: appId },
  }
  const device = {
    id: 'eui-0000000000000002',
    appId,
    joinEui: '0000000000000001',
    devEui: '0000000000000002',
    cac: 'O22322',
  }

  const joinEui = device.joinEui.match(/.{1,2}/g).join(' ')
  const devEui = device.devEui.match(/.{1,2}/g).join(' ')

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, user.ids.user_id)
  })

  beforeEach(() => {
    interceptDeviceRepo(appId)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
  })

  it('succeeds registering a device via a qr code', () => {
    cy.intercept('POST', `/api/v3/applications/${appId}/devices`).as('registerDevice')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByTestId('error-notification').should('not.exist')

    // Open qr modal and scan.
    cy.findByRole('button', { name: /Scan end device QR code/g }).click()
    cy.findByRole('heading', { name: 'Scan end device QR code' }).should('be.visible')
    cy.findByText('Please scan the QR code to continue.').should('exist')
    cy.findByTestId('webcam-feed').should('be.visible')
    cy.findByText('Found QR code data').should('be.visible')
    cy.findByText('Apply').should('not.be.disabled').click()

    // Display scanned data in form.
    cy.findByText('QR code scanned successfully').should('be.visible')

    cy.findByLabelText('End device brand')
      .parents('.select__control')
      .next('input')
      .should('have.value', 'test-brand-otaa')

    selectDevice({
      brand_id: 'test-brand-otaa',
      model_id: 'test-model3',
      hw_version: '2.0',
      fw_version: '1.0.1',
      band_id: 'EU_863_870',
    })

    // End device registration.
    cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
    cy.findByLabelText('JoinEUI').should('have.value', joinEui).should('be.disabled')
    cy.findByLabelText('DevEUI').should('have.value', devEui).should('be.disabled')
    cy.findByRole('button', { name: 'Reset' }).should('be.disabled')
    cy.findByLabelText('AppKey').type(generateHexValue(32))
    cy.findByLabelText('End device ID').should('have.value', `eui-${device.devEui}`)

    cy.findByRole('button', { name: 'Register end device' }).click()

    cy.wait('@registerDevice')
      .location('pathname')
      .should(
        'eq',
        `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/eui-${device.devEui}`,
      )
  })

  it('succeeds scanning again', () => {
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByTestId('error-notification').should('not.exist')

    // Open qr modal and scan.
    cy.findByRole('button', { name: /Scan end device QR code/g }).click()
    cy.findByRole('heading', { name: 'Scan end device QR code' }).should('be.visible')
    cy.findByText('Please scan the QR code to continue.').should('exist')
    cy.findByTestId('webcam-feed').should('be.visible')
    cy.findByText('Found QR code data').should('be.visible')
    cy.findByRole('button', { name: /Scan again/g }).click()
    cy.findByTestId('webcam-feed').should('be.visible')
  })

  it('succeeds resetting QR data', () => {
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByTestId('error-notification').should('not.exist')

    // Open qr modal and scan.
    cy.findByRole('button', { name: /Scan end device QR code/g }).click()
    cy.findByRole('heading', { name: 'Scan end device QR code' }).should('be.visible')
    cy.findByText('Please scan the QR code to continue.').should('exist')
    cy.findByTestId('webcam-feed').should('be.visible')
    cy.findByText('Found QR code data').should('be.visible')
    cy.findByText('Apply').should('not.be.disabled').click()

    // Display scanned data in form.
    cy.findByText('QR code scanned successfully').should('be.visible')

    cy.findByLabelText('End device brand')
      .parents('.select__control')
      .next('input')
      .should('have.value', 'test-brand-otaa')

    selectDevice({
      brand_id: 'test-brand-otaa',
      model_id: 'test-model3',
      hw_version: '2.0',
      fw_version: '1.0.1',
      band_id: 'EU_863_870',
    })

    // End device registration.
    cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
    cy.findByLabelText('JoinEUI').should('have.value', joinEui).should('be.disabled')
    cy.findByLabelText('DevEUI').should('have.value', devEui).should('be.disabled')
    cy.findByRole('button', { name: 'Reset' }).should('be.disabled')
    cy.findByLabelText('AppKey').type(generateHexValue(32))
    cy.findByLabelText('End device ID').should('have.value', `eui-${device.devEui}`)

    // Reset QR data.
    cy.findByRole('button', { name: /Reset QR code data/g }).click()
    cy.findByText(
      'Are you sure you want to discard QR code data? The scanned device will not be registered and the form will be reset.',
    ).should('be.visible')
    cy.findAllByRole('button', { name: /Reset QR code data/g })
      .first()
      .click()

    // Check form is reset
    cy.findByLabelText('End device brand')
      .parents('.select__control')
      .next('input')
      .should('have.value', '')
    cy.findByRole('button', { name: /Scan end device QR code/g }).should('be.visible')
  })

  it('succeeds registering a device via claim', () => {
    cy.intercept('POST', '/api/v3/edcs/claim/info', { body: { supports_claiming: true } }).as(
      'claim-info-request',
    )
    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device)).as('claim-request')

    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByTestId('error-notification').should('not.exist')

    // Open qr modal and scan.
    cy.findByRole('button', { name: /Scan end device QR code/g }).click()
    cy.findByRole('heading', { name: 'Scan end device QR code' }).should('be.visible')
    cy.findByText('Please scan the QR code to continue.').should('exist')
    cy.findByTestId('webcam-feed').should('be.visible')
    cy.findByText('Found QR code data').should('be.visible')
    cy.findByText('Apply').should('not.be.disabled').click()

    cy.findByLabelText('End device brand')
      .parents('.select__control')
      .next('input')
      .should('have.value', 'test-brand-otaa')

    selectDevice({
      brand_id: 'test-brand-otaa',
      model_id: 'test-model3',
      hw_version: '2.0',
      fw_version: '1.0.1',
      band_id: 'EU_863_870',
    })

    cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')

    // Display scanned data in form.
    cy.findByText('QR code scanned successfully').should('be.visible')
    cy.findByLabelText('DevEUI').should('have.value', devEui).should('be.disabled')
    cy.findByLabelText('Claim authentication code')
      .should('have.value', device.cac)
      .should('be.disabled')

    cy.findByRole('button', { name: 'Register end device' }).click()
    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device))
  })
})
