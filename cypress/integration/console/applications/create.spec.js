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

describe('Application create', () => {
  let user

  before(() => {
    cy.dropAndSeedDatabase()
  })

  beforeEach(() => {
    user = {
      ids: { user_id: 'create-app-test-user' },
      primary_email_address: 'create-app-test-user@example.com',
      password: 'ABCDefg123!',
      password_confirm: 'ABCDefg123!',
    }
  })

  it('displays UI elements in place', () => {
    cy.createUser(user)
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/add`)

    cy.findByText('Add application', { selector: 'h1' }).should('be.visible')
    cy.findByLabelText('Application ID')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'my-new-application')
    cy.findByLabelText('Application name')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'My new application')
    cy.findByLabelText('Description')
      .should('be.visible')
      .and('have.attr', 'placeholder')
      .and('eq', 'Description for my new application')
    cy.findDescriptionByLabelText('Description')
      .should(
        'contain',
        'Optional application description; can also be used to save notes about the application',
      )
      .and('be.visible')
    cy.findByLabelText('Linking')
      .should('be.visible')
      .should('have.attr', 'value', 'true')
    cy.findByLabelText('Network Server address').should('be.visible')
    cy.findDescriptionByLabelText('Network Server address')
      .should('contain', 'Leave empty to link to the Network Server in the same cluster')
      .and('be.visible')
    cy.findByRole('button', { name: 'Create application' }).should('be.visible')
  })

  it('validates before submitting an empty form', () => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/add`)

    cy.findByRole('button', { name: 'Create application' })
      .should('be.visible')
      .click()

    cy.findErrorByLabelText('Application ID')
      .should('contain.text', 'Application ID is required')
      .and('be.visible')
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/applications/add`)
  })

  describe('when has no Application Server in the local cluster', () => {
    beforeEach(() => {
      user = {
        ids: { user_id: 'create-app-no-as-test-user' },
        primary_email_address: 'create-app-no-as-test-user@example.com',
        password: 'ABCDefg123!',
        password_confirm: 'ABCDefg123!',
      }
    })

    beforeEach(() => {
      cy.augmentStackConfig(disableApplicationServer)
    })

    it('displays UI elements in place', () => {
      cy.createUser(user)
      cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
      cy.visit(`${Cypress.config('consoleRootPath')}/applications/add`)

      cy.findByText('Add application', { selector: 'h1' }).should('be.visible')
      cy.findByLabelText('Application ID')
        .should('be.visible')
        .and('have.attr', 'placeholder')
        .and('eq', 'my-new-application')
      cy.findByLabelText('Application name')
        .should('be.visible')
        .and('have.attr', 'placeholder')
        .and('eq', 'My new application')
      cy.findByLabelText('Description')
        .should('be.visible')
        .and('have.attr', 'placeholder')
        .and('eq', 'Description for my new application')
      cy.findDescriptionByLabelText('Description')
        .should(
          'contain',
          'Optional application description; can also be used to save notes about the application',
        )
        .and('be.visible')
      cy.findByLabelText('Linking').should('not.exist')
      cy.findByLabelText('Network Server address').should('not.exist')
      cy.findByRole('button', { name: 'Create application' }).should('be.visible')
    })
  })
})
