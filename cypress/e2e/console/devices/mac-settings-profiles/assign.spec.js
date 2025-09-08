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

describe('Device MAC settings profile assignment', () => {
  const userId = 'device-mac-profile-test-user'
  const user = {
    ids: { user_id: userId },
    primary_email_address: 'device-mac-profile-test-user@example.com',
    password: 'ABCDefg123!',
    password_confirm: 'ABCDefg123!',
  }

  const applicationId = 'app-device-mac-profile-test'
  const application = {
    ids: { application_id: applicationId },
    name: 'Application for device MAC profile test',
  }

  const macSettingsProfiles = [
    {
      ids: {
        application_ids: { application_id: applicationId },
        profile_id: 'compatible-profile',
      },
      mac_settings: {
        supports_32_bit_f_cnt: true,
        rx1_delay: 1,
      },
    },
    {
      ids: {
        application_ids: { application_id: applicationId },
        profile_id: 'another-profile',
      },
      mac_settings: {
        supports_32_bit_f_cnt: false,
        rx1_delay: 5,
      },
    },
  ]

  const ns = {
    end_device: {
      frequency_plan_id: 'EU_863_870_TTN',
      lorawan_phy_version: 'PHY_V1_0_2_REV_B',
      multicast: false,
      supports_join: false,
      lorawan_version: 'MAC_V1_0_2',
      session: {
        keys: {
          f_nwk_s_int_key: {
            key: 'CBFBF585D81A9063A31EA6922EDD6360',
          },
        },
        dev_addr: '270000FC',
      },
      supports_class_c: false,
      supports_class_b: false,
      mac_settings: {
        rx2_data_rate_index: 0,
        rx2_frequency: 869525000,
        rx1_delay: 1,
        rx1_data_rate_offset: 0,
        resets_f_cnt: false,
      },
    },
  }

  let endDeviceId

  before(() => {
    cy.dropAndSeedDatabase()
    cy.createUser(user)
    cy.createApplication(application, userId)
    cy.createMockDeviceAllComponents(applicationId, undefined, { ns }).then(body => {
      endDeviceId = body.end_device.ids.device_id
    })
    macSettingsProfiles.forEach(profile => cy.createMacSettingsProfile(applicationId, profile))
  })

  beforeEach(() => {
    cy.loginConsole({ user_id: user.ids.user_id, password: user.password })
    cy.visit(
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/devices/${endDeviceId}/general-settings`,
    )
  })

  it('displays MAC settings profile selection in device edit form', () => {
    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByLabelText('MAC settings profile').should('be.visible')
        cy.findByRole('button', { name: /Remove/ }).should('be.disabled')
      })
  })

  it('succeeds to assign MAC settings profile to existing device', () => {
    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByLabelText('MAC settings profile').selectOption('compatible-profile')
        cy.findByRole('button', { name: /Remove/ }).should('not.be.disabled')
        cy.findByRole('button', { name: /Save changes/ }).click()
      })

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'End device updated')
  })

  it('succeeds to remove MAC settings profile from device', () => {
    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByLabelText('MAC settings profile').selectOption('compatible-profile')
        cy.findByRole('button', { name: /Remove/ }).click()
        cy.findByLabelText('MAC settings profile').should('have.value', '')
        cy.findByRole('button', { name: /Save changes/ }).click()
      })

    cy.findByTestId('toast-notification-success')
      .should('be.visible')
      .and('contain', 'End device updated')
  })

  it('shows custom MAC settings when no profile is selected', () => {
    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByLabelText('MAC settings profile').should('have.value', '')
        cy.findByRole('button', { name: /Remove/ }).should('be.disabled')
        cy.findByText('Custom MAC settings').click()
        cy.get('#mac-settings').scrollIntoView()
        cy.findByText('Frame counter width').should('be.visible')
      })
  })

  it('hides custom MAC settings when profile is selected', () => {
    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByLabelText('MAC settings profile').selectOption('compatible-profile')
        cy.findByText('Custom MAC settings').should('not.exist')
      })
  })

  it('navigates to create MAC settings profile when clicking create button', () => {
    // Mock empty profiles to show the create button
    cy.intercept('GET', `/api/v3/ns/applications/${applicationId}/mac_settings_profiles*`, {
      statusCode: 200,
      body: {
        mac_settings_profiles: [],
      },
    }).as('getEmptyProfiles')

    cy.reload()
    cy.wait('@getEmptyProfiles')

    cy.get('button[type="submit"]').scrollIntoView()
    cy.findByText('Network layer')
      .should('be.visible')
      .closest('[data-test-id="collapsible-section"]')
      .within(() => {
        cy.findByRole('button', { name: 'Expand' }).click()
        cy.get('button[type="submit"]').scrollIntoView()
        cy.findByRole('button', { name: /Create a new MAC settings profile/ }).click()
      })

    cy.location('pathname').should(
      'eq',
      `${Cypress.config('consoleRootPath')}/applications/${applicationId}/mac-settings-profiles/add`,
    )
  })
})
