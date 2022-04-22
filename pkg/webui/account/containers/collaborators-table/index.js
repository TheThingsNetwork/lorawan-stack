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
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getCollaboratorsList } from '@account/store/actions/collaborators'
import { selectUserId } from '@account/store/selectors/user'
import {
  selectCollaborators,
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorsError,
} from '@account/store/selectors/collaborators'
import { selectSelectedClientId } from '@account/store/selectors/clients'

import style from './collaborators-table.styl'

const m = defineMessages({
  id: 'User / Organization ID',
  addCollaborator: 'Add collaborator',
})

const CollaboratorsTable = props => {
  const { clientId, currentUserId, ...rest } = props

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'ids',
        displayName: m.id,
        render: ids => {
          const isUser = 'user_ids' in ids
          const collaboratorId = getCollaboratorId({ ids })
          if (isUser && collaboratorId === currentUserId) {
            return (
              <span>
                {collaboratorId}{' '}
                <Message className={style.hint} content={sharedMessages.currentUserIndicator} />
              </span>
            )
          }
          return collaboratorId
        },
      },
      {
        name: 'ids',
        displayName: sharedMessages.type,
        render: ids => {
          const isUser = 'user_ids' in ids
          const icon = isUser ? 'user' : 'organization'

          return (
            <span>
              <Icon icon={icon} className={style.collaboratorIcon} />
              <Message content={isUser ? sharedMessages.user : sharedMessages.organization} />
            </span>
          )
        },
      },
      {
        name: 'rights',
        displayName: sharedMessages.rights,
        render: rights => {
          for (let i = 0; i < rights.length; i++) {
            if (rights[i].includes('_ALL')) {
              return <Message content={sharedMessages.all} />
            }
          }

          return <span>{rights.length}</span>
        },
      },
    ]

    return baseHeaders
  }, [currentUserId])

  const baseDataSelector = React.useCallback(
    state => ({
      collaborators: selectCollaborators(state, clientId),
      totalCount: selectCollaboratorsTotalCount(state, clientId),
      fetching: selectCollaboratorsFetching(state),
      error: selectCollaboratorsError(state),
    }),
    [clientId],
  )

  const getItems = React.useCallback(
    filters => getCollaboratorsList('client', clientId, filters),
    [clientId],
  )

  const rowKeySelector = React.useCallback(
    row => `${'user_ids' in row.ids ? 'u' : 'c'}_${getCollaboratorId(row)}`,
    [],
  )

  const getCollaboratorPathPrefix = React.useCallback(
    collaborator =>
      `/${'user_ids' in collaborator.ids ? 'user' : 'organization'}/${getCollaboratorId(
        collaborator,
      )}`,
    [],
  )

  return (
    <FetchTable
      entity="collaborators"
      headers={headers}
      rowKeySelector={rowKeySelector}
      getItemPathPrefix={getCollaboratorPathPrefix}
      addMessage={m.addCollaborator}
      getItemsAction={getItems}
      baseDataSelector={baseDataSelector}
      tableTitle={<Message content={sharedMessages.collaborator} />}
      {...rest}
    />
  )
}

CollaboratorsTable.propTypes = {
  clientId: PropTypes.string.isRequired,
  currentUserId: PropTypes.string.isRequired,
}

export default connect(state => ({
  clientId: selectSelectedClientId(state),
  currentUserId: selectUserId(state),
}))(CollaboratorsTable)
