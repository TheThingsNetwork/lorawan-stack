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

describe('Gateway create', () => {
  const gatewayId = 'test-gateway'
  const frequencyPlanId = 'EU_863_870'
  const userId = 'create-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('displays UI elements in place', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/add`)

    cy.findByText('Add gateway', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Gateway ID').should('be.visible')
    cy.findByLabelText('Gateway EUI').should('be.visible')
    cy.findByLabelText('Gateway name').should('be.visible')
    cy.findByLabelText('Gateway description').should('be.visible')
    cy.findDescriptionByLabelText('Gateway description')
      .should(
        'contain',
        'Optional gateway description; can also be used to save notes about the gateway',
      )
      .and('be.visible')
    cy.findByLabelText('Gateway Server address').should('be.visible')
    cy.findDescriptionByLabelText('Gateway Server address')
      .should('contain', 'The address of the Gateway Server to connect to')
      .and('be.visible')
    cy.findByLabelText('Gateway status').should('exist')
    cy.findDescriptionByLabelText('Gateway status')
      .should('contain', 'The status of this gateway may be publicly displayed')
      .and('be.visible')
    cy.findByRole('button', { name: /Add attributes/ }).should('be.visible')
    cy.findByLabelText('Frequency plan').should('be.visible')
    cy.findByLabelText('Schedule downlink late').should('exist')
    cy.findByLabelText('Duty cycle').should('exist')
    cy.findByTestId('schedule_anytime_delay').should('be.visible')
    cy.findByLabelText('Automatic updates').should('exist')
    cy.findDescriptionByLabelText('Automatic updates')
      .should('contain', 'Gateway can be updated automatically')
      .and('be.visible')
    cy.findByLabelText('Channel').should('be.visible')
    cy.findByRole('button', { name: 'Create gateway' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/add`)

    cy.findByRole('button', { name: 'Create gateway' }).click()

    cy.findErrorByLabelText('Gateway ID')
      .should('contain.text', 'Gateway ID is required')
      .and('be.visible')
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/gateways/add`)
  })

  it('successfully adds gateway', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/add`)
    cy.findByLabelText('Gateway ID').type(gatewayId)
    cy.findByLabelText('Frequency plan').selectOption(frequencyPlanId)

    cy.findByRole('button', { name: 'Create gateway' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('full-error-view').should('not.exist')
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`,
    )
  })
})
