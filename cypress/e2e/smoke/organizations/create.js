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

const organizationCreate = defineSmokeTest('succeeds creating organization', () => {
  const user = {
    ids: { user_id: 'org-create-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'org-create-test-user@example.com',
  }
  cy.createUser(user)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  const organization = {
    organization_id: 'org-create-test',
    name: 'Organization Create Test',
    description: 'Organization used in smoke test to verify organization creation',
  }

  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Organizations/ }).click()
  })
  cy.findByRole('link', { name: /Create organization/ }).click()
  cy.findByLabelText('Organization ID').type(organization.organization_id)
  cy.findByLabelText('Name').type(organization.name)
  cy.findByLabelText('Description').type(organization.description)
  cy.findByRole('button', { name: 'Create organization' }).click()

  cy.location('pathname').should(
    'eq',
    `${Cypress.config('consoleRootPath')}/organizations/${organization.organization_id}`,
  )

  cy.findByTestId('error-notification').should('not.exist')
  cy.findByTestId('full-error-view').should('not.exist')
})

export default [organizationCreate]
