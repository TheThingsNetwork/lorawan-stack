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

import { defineSmokeTest } from '../utils'

const organizationDelete = defineSmokeTest('succeeds deleting an organization', () => {
  const user = {
    ids: { user_id: 'org-delete-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'org-delete-test-user@example.com',
  }
  const organization = {
    ids: { organization_id: 'org-delete-test' },
  }
  cy.createUser(user)
  cy.createOrganization(organization, user.ids.user_id)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Organizations/ }).click()
  })
  cy.findByRole('rowgroup').within(() => {
    cy.findByRole('cell', { name: organization.ids.organization_id }).click()
  })
  cy.findByRole('link', { name: /General settings/ }).click()
  cy.findByRole('button', { name: /Delete organization/ }).click()
  cy.findByTestId('modal-window').within(() => {
    cy.findByRole('button', { name: /Delete organization/ }).click()
  })

  cy.findByTestId('full-error-view').should('not.exist')

  cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/organizations`)
  cy.findByRole('table').within(() => {
    cy.findByRole('cell', { name: organization.ids.organization_id }).should('not.exist')
  })
})

export default [organizationDelete]
