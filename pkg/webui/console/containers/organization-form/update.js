// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import toast from '@ttn-lw/components/toast'
import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import {
  composeContact,
  getAdministrativeContact,
  getTechnicalContact,
} from '@ttn-lw/components/contact-fields/utils'

import Require from '@console/lib/components/require'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'
import diff from '@ttn-lw/lib/diff'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayPurgeEntities,
  mayDeleteOrganization,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '@console/lib/feature-checks'

import { updateOrganization, deleteOrganization } from '@console/store/actions/organizations'

import {
  selectSelectedOrganization,
  selectSelectedOrganizationId,
} from '@console/store/selectors/organizations'
import { selectApiKeysTotalCount } from '@console/store/selectors/api-keys'

import OrganizationForm, { initialValues } from './form'

const m = defineMessages({
  deleteOrg: 'Delete organization',
  updateSuccess: 'Organization updated',
  deleteSuccess: 'Organization deleted',
})

const OrganizationUpdateForm = ({ onDeleteSuccess }) => {
  const [error, setError] = useState(undefined)
  const dispatch = useDispatch()
  const orgId = useSelector(selectSelectedOrganizationId)
  const organization = useSelector(selectSelectedOrganization)

  const apiKeysCount = useSelector(selectApiKeysTotalCount)
  const collaboratorsCount = useSelector(selectCollaboratorsTotalCount)
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationApiKeys, state),
  )
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationCollaborators, state),
  )
  const mayPurgeOrg = useSelector(state => checkFromState(mayPurgeEntities, state))

  const hasApiKeys = apiKeysCount > 0
  const hasAddedCollaborators = collaboratorsCount > 1
  const isPristine = !hasApiKeys && !hasAddedCollaborators

  const shouldConfirmDelete = !isPristine || !mayViewCollaborators || !mayViewApiKeys

  const handleUpdate = useCallback(
    async updated => {
      try {
        setError()

        const {
          _administrative_contact_id,
          _administrative_contact_type,
          _technical_contact_id,
          _technical_contact_type,
        } = updated

        const administrative_contact =
          _administrative_contact_id !== ''
            ? composeContact(_administrative_contact_type, _administrative_contact_id)
            : ''

        const technical_contact =
          _technical_contact_id !== ''
            ? composeContact(_technical_contact_type, _technical_contact_id)
            : ''

        const changed = diff(
          organization,
          { administrative_contact, technical_contact, ...updated },
          {
            exclude: [
              'created_at',
              'updated_at',
              '_administrative_contact_id',
              '_administrative_contact_type',
              '_technical_contact_id',
              '_technical_contact_type',
            ],
          },
        )

        if (technical_contact === '') {
          changed.technical_contact = null
        }
        if (administrative_contact === '') {
          changed.administrative_contact = null
        }
        await dispatch(attachPromise(updateOrganization(orgId, changed)))

        toast({
          title: orgId,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setError(error)
      }
    },
    [dispatch, orgId, organization],
  )

  const handleDelete = useCallback(
    async shouldPurge => {
      try {
        await dispatch(attachPromise(deleteOrganization(orgId, { purge: shouldPurge })))
        toast({
          title: orgId,
          message: m.deleteSuccess,
          type: toast.types.SUCCESS,
        })
        onDeleteSuccess()
      } catch (err) {
        setError(err)
      }
    },
    [dispatch, onDeleteSuccess, orgId],
  )

  const deleteButton = (
    <Require featureCheck={mayDeleteOrganization}>
      <DeleteModalButton
        entityId={orgId}
        entityName={organization.name}
        message={m.deleteOrg}
        onApprove={handleDelete}
        shouldConfirm={shouldConfirmDelete}
        mayPurge={mayPurgeOrg}
      />
    </Require>
  )

  // Add technical and administrative contact to the initial values.
  const { administrative_contact, technical_contact, ...organizationValues } = organization
  const technicalContact = getTechnicalContact(organization)
  const administrativeContact = getAdministrativeContact(organization)
  const composedInitialValues = {
    ...initialValues,
    ...technicalContact,
    ...administrativeContact,
    ...organizationValues,
  }

  return (
    <OrganizationForm
      update
      onSubmit={handleUpdate}
      error={error}
      submitBarItems={deleteButton}
      initialValues={composedInitialValues}
      submitMessage={sharedMessages.saveChanges}
    />
  )
}

OrganizationUpdateForm.propTypes = {
  onDeleteSuccess: PropTypes.func.isRequired,
}

export default OrganizationUpdateForm
