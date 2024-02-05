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

describe('Collaborators', () => {
  const userId = 'main-collab-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'main-collab-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const collaboratorId = 'collab-test-user'
  const collaboratorUser = {
    ids: { user_id: collaboratorId },
    primary_email_address: 'collab-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }
  const organizationId = 'test-collab-org'
  const organization = {
    ids: { organization_id: organizationId },
  }

  const applicationRights = ['RIGHT_APPLICATION_ALL']
  const gatewayRights = ['RIGHT_GATEWAY_ALL']
  const organizationRights = [
    'RIGHT_APPLICATION_ALL',
    'RIGHT_CLIENT_ALL',
    'RIGHT_GATEWAY_ALL',
    'RIGHT_ORGANIZATION_ALL',
  ]

  const generateCollaborator = (entity, type) => {
    const userIds = {
      user_ids: {
        user_id: collaboratorId,
      },
    }

    const orgIds = {
      organization_ids: {
        organization_id: organizationId,
      },
    }

    return {
      collaborator: {
        ids: type === 'user' ? userIds : orgIds,
        rights:
          entity === 'applications'
            ? applicationRights
            : entity === 'gateways'
              ? gatewayRights
              : organizationRights,
      },
    }
  }

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createUser(collaboratorUser)
    cy.createOrganization(organization, collaboratorId)
  })

  describe('Application', () => {
    const applicationId = 'collaborators-test-app'
    const application = { ids: { application_id: applicationId } }
    const entity = 'applications'
    const orgCollaborator = generateCollaborator(entity, 'org')
    const userCollaborator = generateCollaborator(entity, 'user')

    before(() => {
      cy.createApplication(application, userId)
      cy.createCollaborator(entity, applicationId, orgCollaborator)
      cy.createCollaborator(entity, applicationId, userCollaborator)
    })

    it('succeeds editing organization collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/collaborators/organization/${organizationId}`,
      )

      cy.findByLabelText('Grant individual rights').check()
      cy.findByLabelText('Select all').check()

      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`Collaborator rights updated`)
        .should('be.visible')
    })

    it('succeeds deleting organization collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/collaborators/organization/${organizationId}`,
      )

      cy.findByRole('button', { name: /Remove collaborator/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Remove collaborator', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Remove collaborator/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/collaborators`)

      cy.findByText(/Collaborators \(\d+\)/).should('be.visible')
      cy.findByRole('cell', { name: organizationId }).should('not.exist')
    })

    it('succeeds editing user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByLabelText('Grant individual rights').check()
      cy.findByLabelText('Select all').check()

      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`Collaborator rights updated`)
        .should('be.visible')
    })

    it('succeeds deleting user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/applications/${applicationId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByRole('button', { name: /Remove collaborator/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Remove collaborator', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Remove collaborator/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.visit(`${Cypress.config('consoleRootPath')}/applications/${applicationId}/collaborators`)

      cy.findByText(/Collaborators \(\d+\)/).should('be.visible')
      cy.findByRole('cell', { name: collaboratorId }).should('not.exist')
    })
  })

  describe('Gateway', () => {
    const gatewayId = 'collaborators-test-gateway'
    const gateway = { ids: { gateway_id: gatewayId } }
    const entity = 'gateways'
    const orgCollaborator = generateCollaborator(entity, 'org')
    const userCollaborator = generateCollaborator(entity, 'user')

    before(() => {
      cy.createGateway(gateway, userId)
      cy.createCollaborator(entity, gatewayId, orgCollaborator)
      cy.createCollaborator(entity, gatewayId, userCollaborator)
    })

    it('succeeds editing organization collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/gateways/${gatewayId}/collaborators/organization/${organizationId}`,
      )

      cy.findByLabelText('Grant individual rights').check()
      cy.findByLabelText('Select all').check()

      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`Collaborator rights updated`)
        .should('be.visible')
    })

    it('succeeds deleting organization collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/gateways/${gatewayId}/collaborators/organization/${organizationId}`,
      )

      cy.findByRole('button', { name: /Remove collaborator/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Remove collaborator', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Remove collaborator/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/collaborators`)

      cy.findByText(/Collaborators \(\d+\)/).should('be.visible')
      cy.findByRole('cell', { name: organizationId }).should('not.exist')
    })

    it('succeeds editing user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/gateways/${gatewayId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByLabelText('Grant individual rights').check()
      cy.findByLabelText('Select all').check()

      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`Collaborator rights updated`)
        .should('be.visible')
    })

    it('succeeds deleting user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/gateways/${gatewayId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByRole('button', { name: /Remove collaborator/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Remove collaborator', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Remove collaborator/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.visit(`${Cypress.config('consoleRootPath')}/gateways/${gatewayId}/collaborators`)

      cy.findByText(/Collaborators \(\d+\)/).should('be.visible')
      cy.findByRole('cell', { name: collaboratorId }).should('not.exist')
    })
  })

  describe('Organization', () => {
    const orgId = 'collaborators-test-org'
    const org = { ids: { organization_id: orgId } }
    const entity = 'organizations'
    const userCollaborator = generateCollaborator(entity, 'user')

    before(() => {
      cy.createOrganization(org, userId)
      cy.createCollaborator(entity, orgId, userCollaborator)
    })

    it('succeeds editing user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/organizations/${orgId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByLabelText('Grant individual rights').check()
      cy.findByLabelText('Select all').check()

      cy.findByRole('button', { name: 'Save changes' }).click()

      cy.findByTestId('error-notification').should('not.exist')
      cy.findByTestId('toast-notification')
        .should('be.visible')
        .findByText(`Collaborator rights updated`)
        .should('be.visible')
    })

    it('succeeds deleting user collaborator', () => {
      cy.loginConsole({ user_id: userId, password: user.password })
      cy.visit(
        `${Cypress.config(
          'consoleRootPath',
        )}/organizations/${orgId}/collaborators/user/${collaboratorId}`,
      )

      cy.findByRole('button', { name: /Remove collaborator/ }).click()

      cy.findByTestId('modal-window')
        .should('be.visible')
        .within(() => {
          cy.findByText('Remove collaborator', { selector: 'h1' }).should('be.visible')
          cy.findByRole('button', { name: /Remove collaborator/ }).click()
        })

      cy.findByTestId('error-notification').should('not.exist')

      cy.visit(`${Cypress.config('consoleRootPath')}/organizations/${orgId}/collaborators`)

      cy.findByText(/Collaborators \(\d+\)/).should('be.visible')
      cy.findByRole('cell', { name: collaboratorId }).should('not.exist')
    })
  })
})
