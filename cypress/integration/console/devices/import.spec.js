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

describe('End device messaging', () => {
  const userId = 'import-devices-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'view-overview-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const appId = 'import-devices-test-application'
  const application = { ids: { application_id: appId } }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices/import`)
  })

  it('succeeds uploading a device file', () => {
    cy.findByText('Import end devices').should('be.visible')
    cy.findByLabelText('File format').selectOption('The Things Stack JSON')

    const devicesFile = 'successful-devices.json'
    cy.findByLabelText('File').attachFile(devicesFile)
    cy.findByRole('button', { name: 'Import end devices' }).click()
    cy.findByText('0 of 3 (0.00% finished)')
    cy.findByText('Operation finished')
    cy.findByText('3 of 3 (100.00% finished)')
    cy.findByTestId('notification')
      .should('be.visible')
      .findByText('All end devices imported successfully')
      .should('be.visible')
    cy.findByRole('button', { name: 'Proceed to end device list' }).click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/devices`,
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByText('migration-test-device').should('be.visible')
    cy.findByText('some-nice-id').should('be.visible')
    cy.findByText('this-is-test-id').should('be.visible')
  })

  it('fails adding devices with existant ids', () => {
    cy.findByLabelText('File format').selectOption('The Things Stack JSON')
    cy.findByLabelText('File').attachFile('duplicate-devices-a.json')
    cy.findByRole('button', { name: 'Import end devices' }).click()
    cy.findByText('Operation finished').should('be.visible')
    cy.reload()

    cy.findByLabelText('File format').selectOption('The Things Stack JSON')
    cy.findByLabelText('File').attachFile('duplicate-devices-b.json')
    cy.findByRole('button', { name: 'Import end devices' }).click()

    cy.findByText('Operation finished').should('be.visible')
    cy.findByText('3 of 3 (100.00% finished)').should('be.visible')
    cy.findByText('Successfully converted 1 of 3 end devices').should('be.visible')
    cy.findByTestId('notification')
      .should('be.visible')
      .findByText('Not all devices imported successfully')
      .should('be.visible')
    cy.findByText('The registration of the following end devices failed:')
      .should('be.visible')
      .closest('div')
      .within(() => {
        cy.findByText(/ID already taken/).should('be.visible')
        cy.findByText(/EUI already taken/).should('be.visible')
      })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${appId}/devices`)
    cy.findByText(/End devices \(\d+\)/).should('be.visible')
    cy.findByText('some-fail-id').should('not.exist')
  })

  it('succeeds setting lorawan_version, lorawan_phy_version and frequency_plan_id from fallback values', () => {
    const devicesFile = 'freqId-version-phy-device.json'
    const fallbackValues = {
      lorawan_version: 'MAC_V1_0',
      frequency_plan_id: '863-870 MHz',
    }
    cy.findByLabelText('File format').selectOption('The Things Stack JSON')
    cy.findByLabelText('File').attachFile(devicesFile)
    cy.findByLabelText('Frequency plan').selectOption(fallbackValues.frequency_plan_id)
    cy.findByLabelText('LoRaWAN version').selectOption(fallbackValues.lorawan_version)

    cy.findByRole('button', { name: 'Import end devices' }).click()
    cy.findByText('Operation finished').should('be.visible')
    cy.findByText('3 of 3 (100.00% finished)').should('be.visible')
    cy.findByTestId('notification')
      .findByText('All end devices imported successfully')
      .should('be.visible')
    cy.findByRole('button', { name: 'Proceed to end device list' }).click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${appId}/devices`,
    )
    cy.findByTestId('error-notification').should('not.exist')
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/devices/fallback-test-device/general-settings`,
    )

    cy.findByText('Network layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText('Europe 863-870 MHz (SF12 for RX2)').should('be.visible')
        cy.findByText('LoRaWAN Specification 1.0.0').should('be.visible')
      })
    cy.visit(
      `${Cypress.config(
        'consoleRootPath',
      )}/applications/${appId}/devices/fallback-test-nice-id/general-settings`,
    )
    cy.findByText('Network layer', { selector: 'h3' })
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.findByText('Europe 863-870 MHz (SF9 for RX2 - recommended)').should('be.visible')
        cy.findByText('LoRaWAN Specification 1.0.3').should('be.visible')
      })
  })

  it('fails importing device without lorawan_version, lorawan_phy_version and frequency_plan_id', () => {
    const devicesFile = 'no-freqId-version-phy-device.json'
    cy.findByLabelText('File format').selectOption('The Things Stack JSON')
    cy.findByLabelText('File').attachFile(devicesFile)
    cy.findByRole('button', { name: 'Import end devices' }).click()
    cy.findByText('Operation finished').should('be.visible')
    cy.findByText('3 of 3 (100.00% finished)').should('be.visible')
    cy.findByText('Successfully converted 2 of 3 end devices').should('be.visible')
    cy.findByTestId('notification')
      .findByText('Not all devices imported successfully')
      .should('be.visible')
    cy.findByText('The registration of the following end device failed:')
      .should('be.visible')
      .closest('div')
      .within(() => {
        cy.findByText('frequency plan `` not found').should('be.visible')
      })
  })
})
