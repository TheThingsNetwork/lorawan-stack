// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect } from 'react'
import { defineMessages } from 'react-intl'
import PropTypes from 'prop-types'
import { useSelector } from 'react-redux'

import Select from '@ttn-lw/components/select'
import Form, { useFormContext } from '@ttn-lw/components/form'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import { getValuesNormalized } from '@console/containers/gateway-managed-gateway/shared/utils'

import { selectCollaboratorsEntitiesStore } from '@ttn-lw/lib/store/selectors/collaborators'
import { getOrganizationId } from '@ttn-lw/lib/selectors/id'

import { checkFromState, mayViewOrganizationsOfUser } from '@console/lib/feature-checks'

import { getOrganizationsList } from '@console/store/actions/organizations'

import { selectOrganizationEntitiesStore } from '@console/store/selectors/organizations'
import { selectUserId } from '@account/store/selectors/user'

const m = defineMessages({
  showProfilesOf: 'Show profiles of',
  yourself: 'Yourself',
  showProfilesOfTooltip:
    'Managed gateways can be setup with WiFi profiles shared within an organization. This dropdown allows you to select the source of the profiles applicable to this gateway. If you are a collaborator of an organization, you can select the organization to view its shared profiles.',
})

const ShowProfilesSelect = ({ name, ...rest }) => {
  const { values, setFieldValue } = useFormContext()
  const organizations = useSelector(selectOrganizationEntitiesStore)
  const collaborators = useSelector(selectCollaboratorsEntitiesStore)
  const userId = useSelector(selectUserId)

  const mayViewOrganizations = useSelector(state =>
    checkFromState(mayViewOrganizationsOfUser, state),
  )

  const value = getValuesNormalized(name, values)

  useEffect(() => {
    if (!Boolean(value)) {
      const collaboratorIds = Object.keys(collaborators).filter(key => {
        const value = collaborators[key]
        const userIdCheck = value.ids.user_ids?.user_id === userId
        const organizationIdCheck = Boolean(value.ids.organization_ids?.organization_id)
        return userIdCheck || organizationIdCheck
      })
      setFieldValue(
        name,
        collaboratorIds.length && collaboratorIds[0] !== userId ? collaboratorIds[0] : userId,
      )
    }
  }, [collaborators, name, setFieldValue, userId, value])

  const profileOptions = [
    { value: userId, label: m.yourself },
    ...Object.values(organizations)
      .filter(({ ids }) => Boolean(collaborators[getOrganizationId({ ids })]))
      .map(({ ids, name }) => ({
        value: getOrganizationId({ ids }),
        label: name ?? getOrganizationId({ ids }),
      })),
  ]

  const loadData = useCallback(
    async dispatch => {
      if (mayViewOrganizations) {
        dispatch(getOrganizationsList(undefined, ['name']))
      }
    },
    [mayViewOrganizations],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <Form.Field
        name={name}
        title={m.showProfilesOf}
        component={Select}
        options={profileOptions}
        disabled={!Object.keys(organizations).length}
        tooltip={m.showProfilesOfTooltip}
        {...rest}
      />
    </RequireRequest>
  )
}

ShowProfilesSelect.propTypes = {
  name: PropTypes.string.isRequired,
}

export default ShowProfilesSelect
