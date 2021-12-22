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
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 3)
    })
  })

  it('fails adding devices with existant ids', () => {
    const devicesFile = 'failed-devices.json'
    cy.findByLabelText('File format').selectOption('The Things Stack JSON')
    cy.findByLabelText('File').attachFile(devicesFile)
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
        cy.findByText(/an end device with/, /is already registered/).should('be.visible')
      })
  })
})
