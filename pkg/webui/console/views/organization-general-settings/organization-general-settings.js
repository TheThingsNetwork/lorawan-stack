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

import React from 'react'
import bind from 'autobind-decorator'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import DeleteModalButton from '@ttn-lw/console/components/delete-modal-button'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import SubmitBar from '@ttn-lw/components/submit-bar'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import Form from '@ttn-lw/components/form'

import OrganizationForm from '@console/components/organization-form'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'
import Require from '@console/lib/components/require'

import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  mayEditBasicOrganizationInformation,
  mayDeleteOrganization,
} from '@console/lib/feature-checks'

const m = defineMessages({
  deleteOrg: 'Delete organization',
  updateSuccess: 'Organization updated',
})

@withFeatureRequirement(mayEditBasicOrganizationInformation, {
  redirect: ({ orgId }) => `/organizations/${orgId}`,
})
@withBreadcrumb('orgs.single.general-settings', ({ orgId }) => (
  <Breadcrumb
    path={`/organizations/${orgId}/general-settings`}
    content={sharedMessages.generalSettings}
  />
))
class GeneralSettings extends React.PureComponent {
  static propTypes = {
    deleteOrganization: PropTypes.func.isRequired,
    deleteOrganizationSuccess: PropTypes.func.isRequired,
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
    shouldConfirmDelete: PropTypes.bool.isRequired,
    shouldPurge: PropTypes.bool.isRequired,
    updateOrganization: PropTypes.func.isRequired,
  }

  state = {
    error: '',
  }

  formRef = React.createRef()

  @bind
  async handleUpdate(updated) {
    await this.setState({ error: '' })

    const { updateOrganization, orgId, organization: original } = this.props
    const changed = diff(original, updated, ['created_at', 'updated_at'])

    return updateOrganization(orgId, changed)
  }

  @bind
  handleUpdateFailure(error) {
    this.setState({ error })
  }

  @bind
  handleUpdateSuccess() {
    const { orgId, organization } = this.props

    if (this.formRef && this.formRef.current) {
      this.formRef.current.resetForm({ values: organization })
    }

    toast({
      title: orgId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }

  @bind
  async handleDelete() {
    const { orgId, deleteOrganization, deleteOrganizationSuccess } = this.props

    await this.setState({ error: '' })

    try {
      await deleteOrganization(orgId)
      deleteOrganizationSuccess()
    } catch (error) {
      this.setState({ error })
    }
  }

  render() {
    const { organization, orgId, shouldConfirmDelete, shouldPurge } = this.props
    const { error } = this.state

    const initialValues = {
      ids: {
        organization_id: orgId,
      },
      name: organization.name || '',
      description: organization.description || '',
    }

    return (
      <Container>
        <PageTitle title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12}>
            <OrganizationForm
              update
              formRef={this.formRef}
              error={error}
              initialValues={initialValues}
              onSubmit={this.handleUpdate}
              onSubmitSuccess={this.handleUpdateSuccess}
              onSubmitFailure={this.handleUpdateFailure}
            >
              <SubmitBar>
                <Form.Submit message={sharedMessages.saveChanges} component={SubmitButton} />
                <Require featureCheck={mayDeleteOrganization}>
                  <DeleteModalButton
                    entityId={orgId}
                    entityName={organization.name}
                    message={m.deleteOrg}
                    onApprove={this.handleDelete}
                    shouldConfirm={shouldConfirmDelete}
                    shouldPurge={shouldPurge}
                  />
                </Require>
              </SubmitBar>
            </OrganizationForm>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default GeneralSettings
