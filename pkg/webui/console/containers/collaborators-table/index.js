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

import React, { useCallback } from 'react'
import { defineMessages, useIntl } from 'react-intl'
import { connect } from 'react-redux'

import Tag from '@ttn-lw/components/tag'
import TagGroup from '@ttn-lw/components/tag/group'
import Icon from '@ttn-lw/components/icon'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import getByPath from '@ttn-lw/lib/get-by-path'

import { selectUserId } from '@console/store/selectors/logout'

import style from './collaborators-table.styl'

const RIGHT_TAG_MAX_WIDTH = 140

const m = defineMessages({
  id: 'User / Organization ID',
})

const rowKeySelector = row => row._type

const getCollaboratorPathPrefix = collaborator =>
  `/${collaborator._type.startsWith('u') ? 'user' : 'organization'}/${getCollaboratorId(
    collaborator,
  )}`

const CollaboratorsTable = props => {
  const { baseDataSelector, ...restProps } = props
  const intl = useIntl()
  const headers = [
    {
      name: 'ids',
      displayName: m.id,
      sortable: true,
      sortKey: '_id',
      width: 30,
      render: ids => {
        const isUser = 'user_ids' in ids
        const collaboratorId = getCollaboratorId({ ids })
        const icon = isUser ? 'user' : 'organization'
        let userLabel = collaboratorId

        if (isUser && collaboratorId === props.currentUserId) {
          userLabel = (
            <span>
              {collaboratorId}{' '}
              <Message className="tc-subtle-gray" content={sharedMessages.currentUserIndicator} />
            </span>
          )
        }
        return (
          <>
            <Icon icon={icon} className="mr-cs-xs" />
            {userLabel}
          </>
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

  const decoratedBaseDataSelector = useCallback(
    (state, props) => {
      const base = baseDataSelector(state, props)

      // Decorate the base data with a unified id and type that we can sort on.
      base.collaborators = base.collaborators.map(c => {
        const _id =
          getByPath(c, 'ids.user_ids.user_id') ||
          getByPath(c, 'ids.organization_ids.organization_id')

        return {
          ...c,
          _id,
          _type: Boolean(getByPath(c, 'ids.user_ids')) ? `u_${_id}` : `o_${_id}`,
        }
      })

      return base
    },
    [baseDataSelector],
  )

  return (
    <FetchTable
      entity="collaborators"
      headers={headers}
      defaultOrder="_id"
      rowKeySelector={rowKeySelector}
      getItemPathPrefix={getCollaboratorPathPrefix}
      addMessage={sharedMessages.addCollaborator}
      tableTitle={<Message content={sharedMessages.collaborators} />}
      baseDataSelector={decoratedBaseDataSelector}
      handlesSorting
      {...restProps}
    />
  )
}

CollaboratorsTable.propTypes = {
  baseDataSelector: PropTypes.func.isRequired,
  currentUserId: PropTypes.string.isRequired,
}

export default connect(state => ({
  currentUserId: selectUserId(state),
}))(CollaboratorsTable)
