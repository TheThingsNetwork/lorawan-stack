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

import { generateCollaborator } from '../../../support/utils'

describe('Organization general settings', () => {
  const organizationId = 'test-organization'
  const organization = { ids: { organization_id: organizationId } }
  const userId = 'main-organization-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'edit-organization-test-user@example.com',
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
    cy.createOrganization(organization, userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/general-settings`,
    )
  })

  it('succeeds editing organization name and description', () => {
    cy.findByLabelText('Name').type('test-name')
    cy.findByLabelText('Description').type('test-description')

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification')
      .should('be.visible')
      .findByText(`Organization updated`)
      .should('be.visible')
  })

  it('fails adding non-collaborator contact information', () => {
    const entity = 'organizations'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, organizationId, userCollaborator)

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').type('test-non-collab-user')
    cy.findByText('No matching user or organization was found')
  })

  it('suceeds adding contact information', () => {
    const entity = 'organizations'
    const userCollaborator = generateCollaborator(entity, 'user')
    cy.createCollaborator(entity, organizationId, userCollaborator)

    cy.visit(
      `${Cypress.config('consoleRootPath')}/organizations/${organizationId}/general-settings`,
    )

    cy.findByText('Contact information').should('be.visible')
    cy.findByLabelText('Administrative contact').clear()
    cy.findByLabelText('Administrative contact').selectOption(collabUserId)

    cy.findByRole('button', { name: 'Save changes' }).click()

    cy.findByTestId('error-notification').should('not.exist')
    cy.findByTestId('toast-notification').findByText(`Organization updated`).should('be.visible')
  })

  it('succeeds deleting organization', () => {
    cy.findByRole('button', { name: /Delete organization/ }).click()

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion', { selector: 'h1' }).should('be.visible')
        cy.get('input').type(organizationId)
        cy.findByRole('button', { name: /Delete organization/ }).click()
      })

    cy.findByTestId('error-notification').should('not.exist')

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/organizations`)

    cy.findByRole('cell', { name: organizationId }).should('not.exist')
  })
})
