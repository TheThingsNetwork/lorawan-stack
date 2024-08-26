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

import React from 'react'
import { createSelector } from 'reselect'
import { useSelector } from 'react-redux'
import { useIntl } from 'react-intl'

import Tag from '@ttn-lw/components/tag'
import TagGroup from '@ttn-lw/components/tag/group'
import Icon, { IconOrganization, IconUser } from '@ttn-lw/components/icon'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'

import { selectUserId } from '@console/store/selectors/user'

import style from './collaborators-table.styl'

const RIGHT_TAG_MAX_WIDTH = 140

const rowKeySelector = row => row._type

const getCollaboratorPathPrefix = collaborator =>
  `${collaborator._type.startsWith('u') ? 'user' : 'organization'}/${getCollaboratorId(
    collaborator,
  )}`

const CollaboratorsTable = props => {
  const { baseDataSelector, isMember, ...restProps } = props
  const intl = useIntl()
  const userId = useSelector(selectUserId)
  const headers = [
    {
      name: 'ids',
      displayName: sharedMessages.userOrgId,
      sortable: true,
      sortKey: 'id',
      width: 30,
      render: ids => {
        const isUser = 'user_ids' in ids
        const collaboratorId = getCollaboratorId({ ids })
        const icon = isUser ? IconUser : IconOrganization
        let userLabel = collaboratorId

        if (isUser && collaboratorId === userId) {
          userLabel = (
            <span>
              {collaboratorId}{' '}
              <Message
                className="c-text-neutral-light"
                content={sharedMessages.currentUserIndicator}
              />
            </span>
          )
        }
        return (
          <div className="d-flex al-center">
            <Icon icon={icon} className="mr-cs-xs" />
            {userLabel}
          </div>
        )
      },
    },
    {
      name: 'rights',
      sortable: true,
      width: 70,
      displayName: sharedMessages.rights,
      align: 'left',
      render: (rights = []) => {
        if (rights.length === 0) {
          return null
        }
        const tags = rights.map(r => {
          let rightLabel = intl.formatMessage({ id: `enum:${r}` })
          rightLabel = rightLabel.charAt(0).toUpperCase() + rightLabel.slice(1)
          return <Tag className={style.rightTag} content={rightLabel} key={r} />
        })

        return <TagGroup tagMaxWidth={RIGHT_TAG_MAX_WIDTH} tags={tags} />
      },
    },
  ]

  const decoratedBaseDataSelector = createSelector(baseDataSelector, base => {
    const collaborators = base.collaborators.map(c => {
      // Decorate the base data with a unified id and type that we can sort on.
      const _id =
        getByPath(c, 'ids.user_ids.user_id') || getByPath(c, 'ids.organization_ids.organization_id')

      return {
        ...c,
        _id,
        _type: Boolean(getByPath(c, 'ids.user_ids')) ? `u_${_id}` : `o_${_id}`,
      }
    })

    return {
      ...base,
      collaborators,
    }
  })

  return (
    <FetchTable
      entity="collaborators"
      headers={headers}
      defaultOrder="id"
      rowKeySelector={rowKeySelector}
      getItemPathPrefix={getCollaboratorPathPrefix}
      addMessage={isMember ? sharedMessages.addMember : sharedMessages.addCollaborator}
      tableTitle={
        <Message content={isMember ? sharedMessages.members : sharedMessages.collaborators} />
      }
      baseDataSelector={decoratedBaseDataSelector}
      handlesSorting
      {...restProps}
    />
  )
}

CollaboratorsTable.propTypes = {
  baseDataSelector: PropTypes.func.isRequired,
  isMember: PropTypes.bool,
}

CollaboratorsTable.defaultProps = {
  isMember: false,
}

export default CollaboratorsTable
