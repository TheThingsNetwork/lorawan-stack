// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import ModalButton from '../../../components/button/modal-button'
import SubmitBar from '../../../components/submit-bar'
import SubmitButton from '../../../components/submit-button'
import toast from '../../../components/toast'
import Form from '../../../components/form'
import OrganizationForm from '../../components/organization-form'

import IntlHelmet from '../../../lib/components/intl-helmet'
import diff from '../../../lib/diff'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'

const m = defineMessages({
  deleteOrg: 'Delete organization',
  modalWarning:
    'Are you sure you want to delete "{orgName}"? Deleting an organization cannot be undone!',
  updateSuccess: 'Successfully updated organization',
})

@withBreadcrumb('orgs.single.general-settings', function(props) {
  const { orgId } = props

  return (
    <Breadcrumb
      path={`/organizations/${orgId}/general-settings`}
      icon="general_settings"
      content={sharedMessages.generalSettings}
    />
  )
})
class GeneralSettings extends React.PureComponent {
  static propTypes = {
    deleteOrganization: PropTypes.func.isRequired,
    deleteOrganizationSuccess: PropTypes.func.isRequired,
    orgId: PropTypes.string.isRequired,
    organization: PropTypes.organization.isRequired,
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
      this.formRef.current.resetForm(organization)
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
    const { organization, orgId } = this.props
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
        <IntlHelmet title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12}>
            <Message component="h2" content={sharedMessages.generalSettings} />
          </Col>
        </Row>
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
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteOrg}
                  modalData={{
                    message: { values: { orgName: organization.name || orgId }, ...m.modalWarning },
                  }}
                  onApprove={this.handleDelete}
                />
              </SubmitBar>
            </OrganizationForm>
          </Col>
        </Row>
      </Container>
    )
  }
}

export default GeneralSettings
