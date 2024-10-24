// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

describe('Delete Managed Gateway', () => {
  const userId = 'managed-gateway-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'managed-gateway-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const gatewayId = 'test-managed-gateway'
  const gateway = { ids: { gateway_id: gatewayId } }

  const gatewayVersionIds = {
    hardware_version: 'v1.1',
    firmware_version: 'v1.1',
    model_id: 'Managed gateway',
  }

  beforeEach(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createGateway(gateway, userId)

    cy.intercept('GET', `/api/v3/gcs/gateways/managed/${gatewayId}*`, {
      statusCode: 200,
      body: {
        ids: {
          gateway_id: `eui-${gateway.eui}`,
          eui: gateway.eui,
        },
        version_ids: gatewayVersionIds,
      },
    }).as('get-is-gtw-managed')

    cy.intercept('POST', `/api/v3/gcls/claim/info`, {
      statusCode: 200,
      body: {
        supports_claiming: true,
      },
    }).as('get-is-gtw-claimable')

    cy.intercept('DELETE', `/api/v3/gcls/claim/${gatewayId}`, {
      statusCode: 200,
    }).as('unclaim-gtw')

    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}`)
    cy.wait('@get-is-gtw-managed')
    cy.findByRole('heading', { name: 'test-managed-gateway' })
    cy.get('button').contains('Managed gateway').should('be.visible')
  })

  it('succeeds to trigger unclaiming when deleting the gateway from the overview header', () => {
    cy.findByTestId('gateway-overview-menu').should('be.visible').click()
    cy.findByText('Unclaim and delete gateway').should('be.visible').click()
    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: 'Unclaim and delete gateway' }).click()
      })
    cy.wait('@unclaim-gtw')
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'Gateway deleted')
  })

  it('succeeds to trigger unclaiming when deleting the gateway from the general settings', () => {
    cy.get('a').contains('General settings').click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/general-settings`,
    )
    cy.findByText('Basic settings').should('be.visible')
    cy.findByRole('button', { name: 'Unclaim and delete gateway' }).click()
    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.findByRole('button', { name: 'Unclaim and delete gateway' }).click()
      })
    cy.wait('@unclaim-gtw')
    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'Gateway deleted')
  })
})
