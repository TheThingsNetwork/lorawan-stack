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

import { disableApplicationServer } from '../../../support/utils'
import { defineSmokeTest } from '../utils'

const applicationCreate = defineSmokeTest('succeeds creating application', () => {
  const user = {
    ids: { user_id: 'app-create-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'app-create-test-user@example.com',
  }
  cy.augmentStackConfig(disableApplicationServer)
  cy.createUser(user)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  const application = {
    application_id: 'app-create-test-app',
    name: 'Application Create Test',
    description: 'Application used in smoke test to verify application creation',
  }
  cy.get('header').within(() => {
    cy.findByRole('link', { name: /Applications/ }).click()
  })
  cy.findByRole('link', { name: /Create application/ }).click()
  cy.findByLabelText('Application ID').type(application.application_id)
  cy.findByLabelText('Application name').type(application.name)
  cy.findByLabelText('Description').type(application.description)
  cy.findByRole('button', { name: 'Create application' }).click()

  cy.location('pathname').should(
    'eq',
    `${Cypress.config('consoleRootPath')}/applications/${application.application_id}`,
  )
  cy.findByTestId('full-error-view').should('not.exist')
})

export default [applicationCreate]
