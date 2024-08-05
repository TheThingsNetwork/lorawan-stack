// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

describe('Notifications', () => {
  const collabUserId = 'test-collab-user'
  const collabUser = {
    ids: { user_id: collabUserId },
    primary_email_address: 'test-collab-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const application = { ids: { application_id: 'test-application' } }
  const userCollaborator = generateCollaborator('applications', 'user')
  const apiKeyName = 'api-test-key'
  const apiKey = {
    name: apiKeyName,
    rights: ['RIGHT_APPLICATION_ALL'],
  }

  beforeEach(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(collabUser)
    cy.createApplication(application, 'admin')
    cy.createCollaborator('applications', application.ids.application_id, userCollaborator)
    cy.createApiKey('applications', application.ids.application_id, apiKey, key => {
      Cypress.config('apiKeyId', key.id)
    })
    cy.loginConsole({ user_id: 'admin', password: 'admin' })
    cy.visit(`${Cypress.config('consoleRootPath')}/notifications`)
  })

  it('succeeds showing a list of notifications', () => {
    cy.findByTestId('notifications-title').should('be.visible')
    cy.findAllByTestId('notification-list-item').should('have.length', 2)
    cy.findByText(/2 unread notifications/).should('be.visible')
    cy.findByText(/Select a notification to display the content here./).should('be.visible')
  })

  it('succeeds opening a notification', () => {
    cy.findAllByTestId('total-unseen-notifications').should('be.visible').and('have.text', '2')
    cy.findAllByTestId('notification-list-item').should('have.length', 2)
    cy.findByText(/Collaborator of application added or updated/).click()
    cy.findByRole('button', { name: /Archive/ }).should('be.visible')
    cy.findAllByTestId('total-unseen-notifications').should('be.visible').and('have.text', '1')
  })

  it('succeeds marking all notifications as read', () => {
    cy.findByTestId('total-unseen-notifications').should('be.visible')
    cy.findByRole('button', { name: /Mark all as read/ }).click()
    cy.findByTestId('total-unseen-notifications').should('not.exist')
  })

  it('succeeds archiving and unarchiving a notification', () => {
    cy.findByText(/Collaborator of application added or updated/).click()
    cy.findByRole('button', { name: /Archive/ }).click()
    cy.findByText(/See archived messages/).click()
    cy.findByText(/Collaborator of application added or updated/).should('be.visible')
    cy.findByText(/Collaborator of application added or updated/).click()
    cy.findByRole('button', { name: /Unarchive/ }).click()
    cy.findByText(/See all messages/).click()
    cy.findByText(/Collaborator of application added or updated/).should('be.visible')
  })

  it('succeeds displaying no notifications text', () => {
    cy.findByText(/See archived messages/).click()
    cy.findByText(/No notifications yet/).should('be.visible')
    cy.findByText(/Once you archive a notification, it can be viewed here./).should('be.visible')
  })
})
