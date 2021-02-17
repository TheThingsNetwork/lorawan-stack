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

const applicationFeatureToggles = defineSmokeTest(
  'restricts access to restricted content correctly',
  () => {
    const user = {
      ids: { user_id: 'feature-toggle-test-user' },
      primary_email_address: 'test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
      email: 'feature-toggle-test-user@example.com',
    }
    const application = { ids: { application_id: 'feature-toggle-test-app' } }
    const rights = [
      'RIGHT_APPLICATION_DELETE',
      'RIGHT_APPLICATION_DEVICES_READ',
      'RIGHT_APPLICATION_DEVICES_READ_KEYS',
      'RIGHT_APPLICATION_DEVICES_WRITE',
      'RIGHT_APPLICATION_DEVICES_WRITE_KEYS',
      'RIGHT_APPLICATION_INFO',
      'RIGHT_APPLICATION_LINK',
      'RIGHT_APPLICATION_SETTINGS_BASIC',
      'RIGHT_APPLICATION_SETTINGS_COLLABORATORS',
      'RIGHT_APPLICATION_SETTINGS_PACKAGES',
      'RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE',
      'RIGHT_APPLICATION_TRAFFIC_READ',
      'RIGHT_APPLICATION_TRAFFIC_UP_WRITE',
    ]
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.createApplication(application, user.ids.user_id)
    cy.setApplicationCollaborator(application.ids.application_id, user.ids.user_id, rights)
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${application.ids.application_id}`)

    cy.findByTestId('navigation-sidebar').within(() => {
      cy.findByText('API Keys').should('not.be.visible')
      cy.findByText('Collaborators').should('be.visible')
    })

    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${
        application.ids.application_id
      }/api-keys`,
    )
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${application.ids.application_id}`,
    )
  },
)

export default [applicationFeatureToggles]
