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
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'
import { createSelector } from 'reselect'

import { PAGE_SIZES } from '@ttn-lw/constants/page-sizes'

import { IconTrash } from '@ttn-lw/components/icon'
import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  deleteAccessToken,
  deleteAllTokens,
  getAccessTokensList,
} from '@console/store/actions/authorizations'

import { selectTokens, selectTokensTotalCount } from '@console/store/selectors/authorizations'
import { selectUserId } from '@console/store/selectors/user'

const m = defineMessages({
  deleteSuccess: 'Access token invalidated',
  deleteFail: 'There was an error and the access token could not be invalidated',
  deleteButton: 'Invalidate this access token',
  deleteAllSuccess: 'All access tokens invalidated',
  deleteAllFail: 'There was an error and the access tokens could not be invalidated',
  deleteAllButton: 'Invalidate all access tokens',
  expiry: 'Expiry',
})

const TokensTable = () => {
  const userId = useSelector(selectUserId)
  const { clientId } = useParams()
  const tokenIdsSelector = createSelector(selectTokens, tokens => tokens.map(token => token.id))
  const tokenIds = useSelector(tokenIdsSelector)
  const dispatch = useDispatch()

  useBreadcrumbs(
    'user-settings.oauth-client-authorizations.single.access-tokens',
    <Breadcrumb
      path={`/user-settings/authorizations/${clientId}/access-tokens`}
      content={sharedMessages.accessTokens}
    />,
  )

  const handleDeleteToken = React.useCallback(
    async id => {
      try {
        await dispatch(attachPromise(deleteAccessToken(userId, clientId, id)))
        toast({
          title: clientId,
          message: m.deleteSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (err) {
        toast({
          title: clientId,
          message: m.deleteFail,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch, userId, clientId],
  )

  const handleDeleteAllTokens = React.useCallback(async () => {
    try {
      await dispatch(attachPromise(deleteAllTokens(userId, clientId, tokenIds)))
      toast({
        title: clientId,
        message: m.deleteAllSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (err) {
      toast({
        title: clientId,
        message: m.deleteAllFail,
        type: toast.types.ERROR,
      })
    }
  }, [dispatch, userId, clientId, tokenIds])

  const headers = React.useMemo(() => {
    const baseHeaders = [
      {
        name: 'id',
        displayName: sharedMessages.id,
        getValue: row => ({
          id: row.id,
        }),
        render: details => (
          <div style={{ maxWidth: '24rem' }}>
            <SafeInspector
              data={details.id}
              noTransform
              noCopyPopup
              small
              hideable={false}
              isBytes={false}
            />
          </div>
        ),
      },
      {
        name: 'created_at',
        displayName: sharedMessages.created,
        width: '4rem',
        sortable: true,
        render: created_at => <DateTime.Relative value={created_at} />,
      },
      {
        name: 'expires_at',
        displayName: m.expiry,
        width: '4rem',
        render: expires_at => <DateTime.Relative value={expires_at} />,
      },
      {
        name: 'actions',
        displayName: sharedMessages.actions,
        width: '6rem',
        getValue: row => ({
          delete: handleDeleteToken.bind(null, row.id),
        }),
        render: details => (
          <Button
            type="button"
            onClick={details.delete}
            message={m.deleteButton}
            icon={IconTrash}
            secondary
            danger
            small
          />
        ),
      },
    ]
    return baseHeaders
  }, [handleDeleteToken])

  const baseDataSelector = createSelector(
    selectTokens,
    selectTokensTotalCount,
    (tokens, totalCount) => ({
      tokens,
      totalCount,
      mayAdd: false,
      mayLink: false,
    }),
  )

  const getItems = React.useCallback(
    filters => getAccessTokensList(userId, clientId, filters),
    [userId, clientId],
  )

  const deleteAllButton = (
    <Button
      type="button"
      onClick={handleDeleteAllTokens}
      message={m.deleteAllButton}
      icon={IconTrash}
      secondary
      danger
    />
  )

  return (
    <div className="container container--xl grid">
      <div className="item-12 xl:item-20">
        <FetchTable
          entity="tokens"
          defaultOrder="-created_at"
          headers={headers}
          getItemsAction={getItems}
          baseDataSelector={baseDataSelector}
          pageSize={PAGE_SIZES.NORMAL}
          actionItems={deleteAllButton}
          tableTitle={<Message content={sharedMessages.accessTokens} />}
        />
      </div>
    </div>
  )
}

export default TokensTable
