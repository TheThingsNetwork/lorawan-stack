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
import { generateHexValue } from '../../../support/utils'

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
    eui: generateHexValue(16),
    gateway_id: 'gateway-create-test',
    name: 'Gateway Create Test',
    frequency_plan_id: 'EU_863_870',
  }

  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Gateways/ }).click()
  })
  cy.findByRole('link', { name: /Register gateway/ }).click()
  cy.findByLabelText('Gateway EUI').type(gateway.eui)
  cy.findByLabelText('Gateway name').type(gateway.name)
  cy.findByLabelText('Frequency plan').selectOption(gateway.frequency_plan_id)
  cy.findByRole('button', { name: 'Register gateway' }).click()

  cy.location('pathname').should(
    'eq',
    `${Cypress.config('consoleRootPath')}/gateways/eui-${gateway.eui}`,
  )
  cy.findByRole('heading', { name: gateway.name })
  cy.findByTestId('error-notification').should('not.exist')
})

export default [gatewayCreate]
