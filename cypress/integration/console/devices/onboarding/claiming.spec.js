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

import { interceptDeviceRepo, selectDevice } from './utils'

const composeExpectedRequest = ({ joinEui, devEui, cac, id, appId }) => ({
  authenticated_identifiers: {
    join_eui: joinEui.toUpperCase(),
    dev_eui: devEui.toUpperCase(),
    authentication_code: cac,
  },
  target_device_id: id,
  target_application_ids: { application_id: appId },
})

const composeClaimResponse = ({ joinEui, devEui, id, appId }) => ({
  application_ids: { application_id: appId },
  device_id: id,
  dev_eui: devEui,
  join_eui: joinEui,
  dev_addr: '2600ABCD',
})

describe('End device repository claiming', () => {
  const user = {
    ids: { user_id: 'claim-test-user' },
    primary_email_address: 'claim-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const appId = 'claim-test-application'
  const application = {
    ids: { application_id: appId },
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, user.ids.user_id)
  })

  beforeEach(() => {
    cy.intercept('POST', '/api/v3/edcs/claim/info', { body: { supports_claiming: true } })
    interceptDeviceRepo(appId)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/add`)
  })

  it('succeeds claiming a device when using the device repository', () => {
    const device1 = {
      id: 'repo-device01',
      appId,
      joinEui: generateHexValue(16),
      devEui: generateHexValue(16),
      cac: '000000001',
    }

    const device2 = {
      id: 'repo-device02',
      appId,
      joinEui: device1.joinEui,
      devEui: generateHexValue(16),
      cac: '000000002',
    }

    // Select device type via device repository.
    selectDevice({
      brand_id: 'test-brand-otaa',
      model_id: 'test-model3',
      hw_version: '2.0',
      fw_version: '1.0.1',
      band_id: 'EU_863_870',
    })

    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device1)).as('claim-request')

    cy.findByLabelText('Frequency plan').selectOption('EU_863_870_TTN')
    cy.findByLabelText('JoinEUI').type(device1.joinEui)
    cy.findByRole('button', { name: 'Confirm' }).click()

    // Provision first device using claiming flow.
    cy.findByLabelText('DevEUI').type(device1.devEui)
    cy.findByLabelText('Claim authentication code').type(device1.cac)
    cy.findByLabelText('End device ID')
      .should('have.value', `eui-${device1.devEui.toLowerCase()}`)
      .clear()
      .type(device1.id)
    cy.findByLabelText('Register another end device of this type').check()
    cy.findByRole('button', { name: 'Add end device' }).click()
    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device1))
    cy.findByTestId('toast-notification').findByText('End device registered').should('be.visible')
    cy.findByRole('button', { name: 'Add end device' }).should('not.be.disabled')

    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device2)).as('claim-request')

    // Provision second device using claiming flow.
    cy.findByLabelText('DevEUI').type(device2.devEui)
    cy.findByLabelText('Claim authentication code').type(device2.cac)
    cy.findByLabelText('End device ID')
      .should('have.value', `eui-${device2.devEui.toLowerCase()}`)
      .clear()
      .type(device2.id)
    cy.findByLabelText('View registered end device').check()
    cy.findByRole('button', { name: 'Add end device' }).click()

    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device2))

    // Check result.
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/devices/${device2.id}`,
    )
    cy.findByRole('heading', { name: device2.id }).should('be.visible')
    cy.get('#stage').within(() => {
      cy.findByRole('button', { name: 'General settings' }).click()
    })
    cy.findByText('Join settings')
      .parents('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByText('Not registered in this cluster').should('be.visible')
      })
    cy.get('#sidebar').within(() => {
      cy.findByRole('link', { name: /End devices/ }).click()
    })
    cy.findByText(device1.id).should('be.visible')
    cy.findByText(device2.id).should('be.visible')
  })

  it('succeeds claiming a multiple devices using manual type specification', () => {
    const device1 = {
      id: 'manual-device01',
      appId,
      joinEui: generateHexValue(16),
      devEui: generateHexValue(16),
      lorawanVersion: 'MAC_V1_0',
      frequencyPlanId: '863-870 MHz',
      cac: '000000001',
    }

    const device2 = {
      id: 'manual-device02',
      appId,
      joinEui: device1.joinEui,
      devEui: generateHexValue(16),
      lorawanVersion: 'MAC_V1_0_3',
      frequencyPlanId: '863-870 MHz',
      cac: '000000002',
    }

    const device3 = {
      id: 'manual-device03',
      appId,
      joinEui: generateHexValue(16),
      devEui: generateHexValue(16),
      lorawanVersion: 'MAC_V1_1',
      phyVersion: 'PHY_V1_1_REV_A',
      frequencyPlanId: '863-870 MHz',
      cac: '000000003',
    }

    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device1)).as('claim-request')

    // Select type info manually.
    cy.findByLabelText('Enter end device specifics manually').check()
    cy.findByLabelText('Frequency plan').selectOption(device1.frequencyPlanId)
    cy.findByLabelText('LoRaWAN version').selectOption(device1.lorawanVersion)
    cy.findByLabelText('JoinEUI').type(device1.joinEui)
    cy.findByRole('button', { name: 'Confirm' }).click()

    // Provision first device using claiming flow.
    cy.findByText('Show advanced activation, LoRaWAN class and cluster settings').click()
    cy.findByLabelText('Cluster settings').should('be.disabled').and('be.checked')
    cy.findByLabelText('DevEUI').type(device1.devEui)
    cy.findByLabelText('Claim authentication code').type(device1.cac)
    cy.findByLabelText('End device ID').clear().type(device1.id)
    cy.findByLabelText('Register another end device of this type').check()
    cy.findByRole('button', { name: 'Add end device' }).click()
    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device1))
    // Properly wait for the form to finish submitting.
    cy.findByTestId('toast-notification').findByText('End device registered').should('be.visible')
    cy.findByRole('button', { name: 'Add end device' }).should('not.be.disabled')

    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device2)).as('claim-request')

    // Provision second device using claiming flow.
    cy.findByLabelText('LoRaWAN version').selectOption(device2.lorawanVersion)
    cy.findByLabelText('DevEUI').type(device2.devEui)
    cy.findByLabelText('Claim authentication code').type(device2.cac)
    cy.findByLabelText('End device ID').clear().type(device2.id)
    cy.findByRole('button', { name: 'Add end device' }).click()
    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device2))
    // Properly wait for the form to finish submitting.
    cy.findByTestId('toast-notification').findByText('End device registered').should('be.visible')
    cy.findByRole('button', { name: 'Add end device' }).should('not.be.disabled')

    cy.intercept('POST', '/api/v3/edcs/claim', composeClaimResponse(device3)).as('claim-request')

    // Provision third device using claiming flow.
    cy.findByLabelText('LoRaWAN version').selectOption(device3.lorawanVersion)
    cy.findByLabelText('Regional Parameters version').selectOption(device3.phyVersion)
    cy.findByRole('button', { name: 'Reset' }).click()
    cy.findByLabelText('JoinEUI').should('not.be.disabled')
    cy.findByLabelText('JoinEUI').type(`${device3.joinEui}{enter}`)
    cy.findByLabelText('DevEUI').type(device3.devEui)
    cy.findByLabelText('Claim authentication code').type(device3.cac)
    cy.findByLabelText('End device ID').clear().type(device3.id)
    cy.findByLabelText('View registered end device').check()
    cy.findByRole('button', { name: 'Add end device' }).click()
    cy.wait('@claim-request')
      .its('request.body')
      .should('deep.equal', composeExpectedRequest(device3))

    // Check result.
    cy.location('pathname').should(
      'eq',
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/devices/${device3.id.toLowerCase()}`,
    )
    cy.findByRole('heading', { name: `${device3.id.toLowerCase()}` }).should('be.visible')
    cy.findByText('Provisioned on external Join Server').should('be.visible')
    cy.get('#stage').within(() => {
      cy.findByRole('button', { name: 'General settings' }).click()
    })
    cy.findByText('Join settings')
      .parents('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByText('Not registered in this cluster').should('be.visible')
      })
    cy.get('#sidebar').within(() => {
      cy.findByRole('link', { name: /End devices/ }).click()
    })
    cy.findByText(device1.id).should('be.visible')
    cy.findByText(device2.id).should('be.visible')
    cy.findByText(device3.id).should('be.visible')
  })
})
