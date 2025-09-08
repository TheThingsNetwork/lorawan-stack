// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

describe('MAC settings profiles list', () => {
  const userId = 'mac-settings-profiles-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'mac-settings-profiles-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applicationId = 'app-mac-settings-profiles-test'
  const application = {
    ids: { application_id: applicationId },
    name: 'Application for MAC settings profiles test',
  }

  const macSettingsProfiles = [
    {
      ids: {
        application_ids: { application_id: applicationId },
        profile_id: 'test-profile-1',
      },
      mac_settings: {
        supports_32_bit_f_cnt: true,
        rx1_delay: 1,
      },
    },
    {
      ids: {
        application_ids: { application_id: applicationId },
        profile_id: 'test-profile-2',
      },
      mac_settings: {
        supports_32_bit_f_cnt: false,
        rx1_delay: 5,
      },
    },
  ]

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    macSettingsProfiles.forEach(profile => cy.createMacSettingsProfile(applicationId, profile))
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles`,
    )
  })

  it('displays MAC settings profiles list', () => {
    cy.findByText('MAC settings profiles', { selector: 'h1' }).should('be.visible')
    cy.findByText(
      'MAC settings profiles are used to set MAC settings of end devices more conveniently.',
    ).should('be.visible')

    cy.findByText('Profile ID').should('be.visible')
    cy.findByText('Actions').should('be.visible')

    cy.findByText('test-profile-1').should('be.visible')
    cy.findByText('test-profile-2').should('be.visible')
  })

  it('succeeds to navigate to create MAC settings profile', () => {
    cy.findByRole('link', { name: /Create a new MAC settings profile/ }).click()
    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles/add`,
    )
  })

  it('succeeds to edit MAC settings profile', () => {
    cy.findByRole('row', { name: /test-profile-1/ })
      .should('be.visible')
      .within(() => {
        cy.findByRole('button', { name: /Edit/ }).click()
      })

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles/test-profile-1`,
    )
  })

  it('succeeds to delete MAC settings profile', () => {
    cy.findByRole('row', { name: /test-profile-1/ })
      .should('be.visible')
      .within(() => {
        cy.findByRole('button', { name: /Delete/ }).click()
      })

    cy.findByTestId('modal-window')
      .should('be.visible')
      .within(() => {
        cy.findByText('Confirm deletion').should('be.visible')
        cy.findByText('The profile will not be applicable to end devices anymore.').should(
          'be.visible',
        )
        cy.findByRole('button', { name: /Delete/ }).click()
      })

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'MAC settings profile deleted')
  })
})
