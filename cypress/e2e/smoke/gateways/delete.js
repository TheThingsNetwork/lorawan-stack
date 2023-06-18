// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

const gatewayDelete = defineSmokeTest('succeeds deleting a gateway', () => {
  const user = {
    ids: { user_id: 'gtw-delete-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'gtw-delete-test-user@example.com',
  }
  cy.createUser(user)
  const gateway = {
    ids: { gateway_id: 'gtw-test-delete' },
  }
  cy.createGateway(gateway, user.ids.user_id)

  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))
  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Gateways/ }).click()
  })
  cy.findByRole('rowgroup').within(() => {
    cy.findByRole('cell', { name: gateway.ids.gateway_id }).click()
  })
  cy.findByRole('link', { name: /General settings/ }).click()
  cy.findByRole('button', { name: /Delete gateway/ }).click()
  cy.findByTestId('modal-window').within(() => {
    cy.findByRole('button', { name: /Delete gateway/ }).click()
  })

  cy.findByTestId('full-error-view').should('not.exist')

  cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/gateways`)
  cy.findByRole('table').within(() => {
    cy.findByRole('cell', { name: gateway.ids.gateway_id }).should('not.exist')
  })
})

export default [gatewayDelete]
