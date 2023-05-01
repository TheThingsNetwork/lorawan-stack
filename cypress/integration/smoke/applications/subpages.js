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

const applicationSubpages = defineSmokeTest('check all application sub-pages', () => {
  const user = {
    ids: { user_id: 'app-sub-pages-test-user' },
    primary_email_address: 'test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
    email: 'app-sub-pages-user@example.com',
  }
  const application = { ids: { application_id: 'integration-test-application' } }
  cy.createUser(user)
  cy.createApplication(application, user.ids.user_id)
  cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  cy.visit(Cypress.config('consoleRootPath'))

  cy.get('header').within(() => {
    cy.findAllByRole('link', { name: /Applications/ })
      .first()
      .click()
  })
  cy.findByRole('cell', { name: application.ids.application_id }).click()

  cy.findAllByRole('link', { name: /End devices/ })
    .first()
    .click()
  cy.findByText('End devices (0)').should('be.visible')
  cy.findByRole('link', { name: /Register end device/ })
    .should('be.visible')
    .click()
  cy.findByRole('heading', { name: 'Register end device' }).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')
  cy.go('back')

  cy.findByRole('link', { name: /Import end devices/ })
    .should('be.visible')
    .click()
  cy.findByRole('heading', { name: /Import end devices/ }).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /Live data/ }).click()
  cy.findByText(/Waiting for events from/).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('button', { name: /Payload formatters/ }).click()

  cy.findByRole('link', { name: /Uplink/ }).click()
  cy.findByLabelText('Formatter type').should('be.visible')
  cy.findByRole('button', { name: 'Save changes' })
  cy.findByRole('link', { name: /Downlink/ }).click()
  cy.findByLabelText('Formatter type').should('be.visible')
  cy.findByRole('button', { name: 'Save changes' })

  cy.findByRole('button', { name: /Integrations/ }).click()

  cy.findByRole('link', { name: /MQTT/ }).click()
  cy.findByText('Connection information').should('be.visible')
  cy.findByText('MQTT server host').should('be.visible')
  cy.findByText('Connection credentials').should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /Webhooks/ }).click()
  cy.findByText('Webhooks (0)').should('be.visible')
  cy.findByText('No items found').should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /Pub\/Subs/ }).click()
  cy.findByText('Pub/Subs (0)').should('be.visible')
  cy.findByText('No items found').should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /LoRa Cloud/ }).click()
  cy.findByText('Geolocation')
    .closest('[data-test-id="collapsible-section"]')
    .within(() => {
      cy.findByRole('button', { name: 'Expand' }).click()
      cy.findByLabelText('Token').should('be.visible')
      cy.findByRole('button', { name: 'Set token' }).should('be.visible')
      cy.findByRole('button', { name: 'Collapse' }).click()
      cy.findByTestId('error-notification').should('not.exist')
    })
  cy.findByRole('heading', { name: 'Device & Application Services' })
    .closest('[data-test-id="collapsible-section"]')
    .within(() => {
      cy.findByRole('button', { name: 'Expand' }).click()
      cy.findByLabelText('Token').should('be.visible')
      cy.findByRole('button', { name: 'Set token' }).should('be.visible')
      cy.findByRole('button', { name: 'Collapse' }).click()
      cy.findByTestId('error-notification').should('not.exist')
    })
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /Collaborators/ }).click()
  cy.findByText('Collaborators (1)').should('be.visible')
  cy.findByRole('link', { name: /Add collaborator/ }).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /API keys/ }).click()
  cy.findByText('API keys (0)').should('be.visible')
  cy.findByRole('link', { name: /Add API key/ }).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')

  cy.findByRole('link', { name: /General settings/ }).click()
  cy.findByLabelText('Application ID').should('be.visible')
  cy.findByRole('button', { name: /Save changes/ }).should('be.visible')
  cy.findByTestId('error-notification').should('not.exist')
})

export default [applicationSubpages]
