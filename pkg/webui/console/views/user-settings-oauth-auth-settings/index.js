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

import React from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages, useIntl } from 'react-intl'
import { useNavigate, useParams } from 'react-router-dom'

import { IconTrash } from '@ttn-lw/components/icon'
import toast from '@ttn-lw/components/toast'
import ModalButton from '@ttn-lw/components/button/modal-button'
import DataSheet from '@ttn-lw/components/data-sheet'
import Tag from '@ttn-lw/components/tag'
import TagGroup from '@ttn-lw/components/tag/group'

import DateTime from '@ttn-lw/lib/components/date-time'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { deleteAuthorization } from '@console/store/actions/authorizations'

import { selectSelectedAuthorization } from '@console/store/selectors/authorizations'

const RIGHT_TAG_MAX_WIDTH = 140

const m = defineMessages({
  deleteButton: 'Revoke authorization',
  deleteSuccess: 'This authorization was successfully revoked',
  deleteFailure: 'There was an error and this authorization could not be revoked',
  deleteMessage:
    'Are you sure you want to unauthorize this client? The client will not be able to perform any actions on your behalf if the authorization is revoked. You can always choose to authorize the client again if wished.',
})

const AuthorizationSettings = () => {
  const { clientId } = useParams()
  const authorization = useSelector(state => selectSelectedAuthorization(state, clientId))
  const { client_ids, user_ids, rights, created_at, updated_at } = authorization
  const { client_id } = client_ids
  const { user_id } = user_ids
  const intl = useIntl()
  const navigate = useNavigate()
  const dispatch = useDispatch()

  const handleDeleteAuthorization = React.useCallback(async () => {
    try {
      await dispatch(attachPromise(deleteAuthorization(user_id, client_id)))
      toast({
        title: client_id,
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
      navigate('/user-settings/authorizations')
    } catch (err) {
      toast({
        title: client_id,
        message: m.deleteFailure,
        type: toast.types.ERROR,
      })
    }
  }, [dispatch, client_id, user_id, navigate])

  const tags = rights.map(r => {
    let rightLabel = intl.formatMessage({ id: `enum:${r}` })
    rightLabel = rightLabel.charAt(0).toUpperCase() + rightLabel.slice(1)
    return <Tag content={rightLabel} key={r} />
  })

  const sheetData = React.useMemo(
    () => [
      {
        header: sharedMessages.generalInformation,
        items: [
          {
            key: sharedMessages.oauthClientId,
            value: user_id,
            type: 'code',
            sensitive: false,
          },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
          { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
          {
            key: sharedMessages.rights,
            value: <TagGroup tagMaxWidth={RIGHT_TAG_MAX_WIDTH} tags={tags} />,
            sensitive: false,
          },
        ],
      },
      {
        header: sharedMessages.actions,
        items: [
          {
            value: (
              <ModalButton
                modalData={{
                  message: m.deleteMessage,
                }}
                onApprove={handleDeleteAuthorization}
                message={m.deleteButton}
                type="button"
                icon={IconTrash}
                danger
              />
            ),
          },
        ],
      },
    ],
    [created_at, user_id, updated_at, tags, handleDeleteAuthorization],
  )

  return (
    <div className="container container--xl grid">
      <div className="item-12">
        <DataSheet data={sheetData} />
      </div>
    </div>
  )
}

export default AuthorizationSettings
