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
import { connect, useDispatch } from 'react-redux'
import { defineMessages } from 'react-intl'
import { bindActionCreators } from 'redux'

import Icon from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'

import { getCollaboratorId } from '@ttn-lw/lib/selectors/id'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { getCollaboratorsList, deleteCollaborator } from '@account/store/actions/collaborators'

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
  deleteCollaboratorError: 'There was an error and the collaborator could not be deleted',
  deleteOnlyCollaboratorError:
    'This collaborator could not be deleted because every client needs at least one collaborator with all rights',
  removeButtonMessage: 'Remove this collaborator',
  removeYourselfMessage: 'Remove yourself as a collaborator',
})

const CollaboratorsTable = props => {
  const { clientId, currentUserId, handleDeleteCollaborator, ...rest } = props
  const dispatch = useDispatch()

  const deleteCollaborator = React.useCallback(
    async ids => {
      const collaboratorType = 'user_ids' in ids ? 'user' : 'organization'
      const collaborator_ids = {
        [`${collaboratorType}_ids`]: {
          [`${collaboratorType}_id`]: getCollaboratorId({ ids }),
        },
      }
      const updatedCollaborator = {
        ids: collaborator_ids,
      }

      try {
        await handleDeleteCollaborator(updatedCollaborator)
        toast({
          message: sharedMessages.collaboratorDeleteSuccess,
          type: toast.types.SUCCESS,
        })
        dispatch(getCollaboratorsList('client', clientId))
      } catch (error) {
        const isOnlyCollaborator = error.details[0].name === 'client_needs_collaborator'

        toast({
          message: isOnlyCollaborator ? m.deleteOnlyCollaboratorError : m.deleteCollaboratorError,
          type: toast.types.ERROR,
        })
      }
    },
    [clientId, dispatch, handleDeleteCollaborator],
  )

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
      {
        name: 'actions',
        displayName: sharedMessages.actions,
        getValue: row => ({
          id: row.ids,
          delete: deleteCollaborator.bind(null, row.ids),
        }),
        render: details => {
          const isUser = 'user_ids' in details.id
          const collaboratorId = getCollaboratorId({ ids: details.id })
          const isYou = isUser && collaboratorId === currentUserId

          return (
            <Button
              type="button"
              onClick={details.delete}
              message={isYou ? m.removeYourselfMessage : m.removeButtonMessage}
              icon="delete"
              danger
            />
          )
        },
      },
    ]

    return baseHeaders
  }, [currentUserId, deleteCollaborator])

  const baseDataSelector = React.useCallback(
    state => ({
      collaborators: selectCollaborators(state, clientId),
      totalCount: selectCollaboratorsTotalCount(state, clientId),
      fetching: selectCollaboratorsFetching(state),
      error: selectCollaboratorsError(state),
      mayLink: false,
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
  handleDeleteCollaborator: PropTypes.func.isRequired,
}

export default connect(
  state => ({
    clientId: selectSelectedClientId(state),
    currentUserId: selectUserId(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        handleDeleteCollaborator: attachPromise(deleteCollaborator),
      },
      dispatch,
    ),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    handleDeleteCollaborator: patch =>
      dispatchProps.handleDeleteCollaborator(stateProps.clientId, patch),
  }),
)(CollaboratorsTable)
