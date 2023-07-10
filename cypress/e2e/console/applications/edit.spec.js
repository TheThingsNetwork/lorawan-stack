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

import { generateCollaborator } from '../../../support/utils'

describe('Application general settings', () => {
  const applicationId = 'test-application'
  const application = { ids: { application_id: applicationId } }
  const userId = 'main-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-application-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const collabUserId = 'test-collab-user'
  const collabUser = {
    ids: { user_id: collabUserId },
    primary_email_address: 'test-collab-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createUser(collabUser)
    cy.createApplication(application, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
  })

  it('succeeds editing application name and description', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)

    cy.findByLabelText('Name').type('test-name')
    cy.findByLabelText('Description').type('test-description')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Application updated`)
      .should('be.visible')
  })

  it('succeeds adding application attributes', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)

    cy.findByRole('button', { name: /Add attributes/ }).click()

    cy.get(`[name="attributes[0].key"]`).type('application-test-key')
    cy.get(`[name="attributes[0].value"]`).type('application-test-value')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Application updated`)
      .should('be.visible')
  })

  it('suceeds at changing skip payload crypto', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)

    cy.findByText('Skip payload encryption and decryption').click()

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`Application updated`).should('be.visible')
  })

  it('fails adding non-collaborator contact information', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').type(collabUserId)
    cy.findByText('No matching user or organization was found')
  })

  it('suceeds adding contact information', () => {
    const entity = 'applications'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, applicationId, userCollaborator)

    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').selectOption(collabUserId)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`Application updated`).should('be.visible')
  })

  it('succeeds deleting application', () => {
    cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/general-settings`)
    cy.findByRole('button', { name: /Delete application/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.get('input').type(applicationId)
        cy.findByRole('button', { name: /Delete application/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/applications`)

    cy.findByRole('cell', { name: applicationId }).should('not.exist')
  })
})
