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

describe('Applications list', () => {
  const userId = 'list-apps-test-user'
  const user = {
    ids: { user_id: 'list-apps-test-user' },
    primary_email_address: 'list-apps-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applications = [
    {
      application_server_address: window.location.hostname,
      ids: {
        application_id: 'xyz-test-app',
      },
      name: 'Application Test Name',
      description: 'Application Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
    {
      application_server_address: window.location.hostname,
      ids: {
        application_id: 'some-test-app',
      },
      name: 'Application Test Name',
      description: 'Application Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
    {
      application_server_address: window.location.hostname,
      ids: {
        application_id: 'other-test-app',
      },
      name: 'Application Test Name',
      description: 'Application Test Description',
      join_server_address: window.location.hostname,
      network_server_address: window.location.hostname,
    },
    {
      application_server_address: 'tti.staging1.cloud.thethings.industries',
      ids: {
        application_id: 'other-cluster-test-app',
      },
      name: 'Application Test Name',
      description: 'Application Test Description',
      join_server_address: 'tti.staging1.cloud.thethings.industries',
      network_server_address: 'tti.staging1.cloud.thethings.industries',
    },
  ]

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(applications[0], userId)
    cy.createApplication(applications[1], userId)
    cy.createApplication(applications[2], userId)
    cy.createApplication(applications[3], userId)
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(`${Cypress.config('consoleRootPath')}/applications`)
  })

  it('succeeds searching by application id', () => {
    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 4)
    })
    cy.findByRole('cell', { name: applications[0].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[1].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[2].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[3].ids.application_id }).should('be.visible')

    cy.findByTestId('search-input').as('searchInput')
    cy.get('@searchInput').type('xyz')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: applications[0].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[1].ids.application_id }).should('not.exist')
    cy.findByRole('cell', { name: applications[2].ids.application_id }).should('not.exist')
    cy.findByRole('cell', { name: applications[3].ids.application_id }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('some')

    cy.findByRole('rowgroup').within(() => {
      cy.findByRole('row').should('have.length', 1)
    })
    cy.findByRole('cell', { name: applications[0].ids.application_id }).should('not.exist')
    cy.findByRole('cell', { name: applications[1].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[2].ids.application_id }).should('not.exist')
    cy.findByRole('cell', { name: applications[3].ids.application_id }).should('not.exist')

    cy.get('@searchInput').clear()
    cy.get('@searchInput').type('test-app')

    cy.findByRole('rowgroup').within(() => {
      cy.findAllByRole('row').should('have.length', 4)
    })
    cy.findByRole('cell', { name: applications[0].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[1].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[2].ids.application_id }).should('be.visible')
    cy.findByRole('cell', { name: applications[3].ids.application_id }).should('be.visible')
  })

  it('succeeds disabling click on applications that are on another cluster', () => {
    cy.findByText(applications[3].ids.application_id).click()
    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/applications`)
    cy.findByTestId('full-error-view').should('not.exist')
  })

  it('succeeds showing "Other cluster" status on applications that are on another cluster', () => {
    cy.findByText(applications[3].ids.application_id)
      .closest('[role="row"]')
      .within(() => {
        cy.findByText('Other cluster').should('be.visible')
      })
  })

  it('succeeds redirecting when manually accessing devices that are on another cluster', () => {
    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applications[3].ids.application_id}`,
    )

    cy.location('pathname').should('eq', `${Cypress.config('consoleRootPath')}/applications`)
    cy.findByTestId('full-error-view').should('not.exist')
    cy.findByText(
      'The application you attempted to visit is registered on a different cluster and needs to be accessed using its host Console.',
    )
  })
})
