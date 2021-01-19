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

const gatewayCreate = defineSmokeTest('succeeds creating gateway', () => {
  const user = {
    ids: { user_id: 'gateway-create-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'gateway-create-test-user@example.com',
  }
  cy.createUser(user)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  const gateway = {
    gateway_id: 'gateway-create-test',
    name: 'Gateway Create Test',
    description: 'Gateway used in smoke test to verify gateway creation',
    gateway_server_address: 'test-address',
    frequency_plan_id: 'EU_863_870',
  }

  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Gateways/ }).click()
  })
  cy.findByRole('link', { name: /Add gateway/ }).click()
  cy.findByLabelText('Gateway ID').type(gateway.gateway_id)
  cy.findByLabelText('Gateway name').type(gateway.name)
  cy.findByLabelText('Gateway description').type(gateway.description)
  cy.findByLabelText('Gateway Server address').type(gateway.gateway_server_address)
  cy.findByLabelText('Frequency plan').selectOption(gateway.frequency_plan_id)
  cy.findByRole('button', { name: 'Create gateway' }).click()

  cy.location('pathname').should(
    'eq',
    `${Cypress.config('consoleRootPath')}/gateways/${gateway.gateway_id}`,
  )

  cy.findByTestId('error-notification').should('not.exist')
  cy.findByTestId('full-error-view').should('not.exist')
})

export default [gatewayCreate]
