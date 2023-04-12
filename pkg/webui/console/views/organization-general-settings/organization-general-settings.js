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
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
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

const GeneralSettings = props => {
  const {
    organization,
    orgId,
    shouldConfirmDelete,
    mayPurge,
    updateOrganization,
    deleteOrganization,
    deleteOrganizationSuccess,
    mayEditBasicInformation,
  } = props

  useBreadcrumbs(
    'orgs.single.general-settings',
    <Breadcrumb
      path={`/organizations/${orgId}/general-settings`}
      content={sharedMessages.generalSettings}
    />,
  )

  const [error, setError] = React.useState('')
  const formRef = React.useRef()

  const initialValues = React.useMemo(
    () => ({
      ids: {
        organization_id: orgId,
      },
      name: organization.name || '',
      description: organization.description || '',
    }),
    [orgId, organization.description, organization.name],
  )

  const handleUpdate = React.useCallback(
    updated => {
      setError('')

      const changed = diff(organization, updated, ['created_at', 'updated_at'])

      return updateOrganization(orgId, changed)
    },
    [orgId, organization, updateOrganization],
  )
  const handleUpdateFailure = React.useCallback(error => setError(error), [])
  const handleUpdateSuccess = React.useCallback(() => {
    toast({
      title: orgId,
      message: m.updateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [orgId])
  const handleDelete = React.useCallback(
    async shouldPurge => {
      try {
        await deleteOrganization(orgId, shouldPurge)
        deleteOrganizationSuccess()
      } catch (err) {
        setError(err)
      }
    },
    [deleteOrganization, deleteOrganizationSuccess, orgId],
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.generalSettings} />
      <Row>
        <Col lg={8} md={12}>
          <OrganizationForm
            update
            formRef={formRef}
            error={error}
            initialValues={initialValues}
            onSubmit={handleUpdate}
            onSubmitSuccess={handleUpdateSuccess}
            onSubmitFailure={handleUpdateFailure}
            mayEditBasicInformation={mayEditBasicInformation}
          >
            <SubmitBar>
              <Form.Submit message={sharedMessages.saveChanges} component={SubmitButton} />
              <Require featureCheck={mayDeleteOrganization}>
                <DeleteModalButton
                  entityId={orgId}
                  entityName={organization.name}
                  message={m.deleteOrg}
                  onApprove={handleDelete}
                  shouldConfirm={shouldConfirmDelete}
                  mayPurge={mayPurge}
                />
              </Require>
            </SubmitBar>
          </OrganizationForm>
        </Col>
      </Row>
    </Container>
  )
}

GeneralSettings.propTypes = {
  deleteOrganization: PropTypes.func.isRequired,
  deleteOrganizationSuccess: PropTypes.func.isRequired,
  mayEditBasicInformation: PropTypes.bool.isRequired,
  mayPurge: PropTypes.bool.isRequired,
  orgId: PropTypes.string.isRequired,
  organization: PropTypes.organization.isRequired,
  shouldConfirmDelete: PropTypes.bool.isRequired,
  updateOrganization: PropTypes.func.isRequired,
}

export default withFeatureRequirement(mayEditBasicOrganizationInformation, {
  redirect: ({ orgId }) => `/organizations/${orgId}`,
})(GeneralSettings)
