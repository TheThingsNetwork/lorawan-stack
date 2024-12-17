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

import React, { useCallback, useMemo, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import { GATEWAY } from '@console/constants/entities'
import tts from '@console/api/tts'

import { IconMenu2, IconStar, IconStarFilled } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'
import Dropdown from '@ttn-lw/components/dropdown'

import Message from '@ttn-lw/lib/components/message'

import GatewayConnection from '@console/containers/gateway-connection'
import DeleteEntityHeaderModal from '@console/containers/delete-entity-header-modal'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'
import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'

import { checkFromState, mayDeleteGateway } from '@console/lib/feature-checks'

import {
  ADD_BOOKMARK_BASE,
  addBookmark,
  DELETE_BOOKMARK_BASE,
  deleteBookmark,
} from '@console/store/actions/user-preferences'

import { selectUser } from '@console/store/selectors/user'
import { selectBookmarksList } from '@console/store/selectors/user-preferences'
import { selectSelectedGatewayClaimable } from '@console/store/selectors/gateways'

import style from './gateway-overview-header.styl'

const m = defineMessages({
  addBookmarkFail: 'There was an error and the gateway could not be bookmarked',
  duplicateGateway: 'Duplicate gateway',
  removeBookmarkFail: 'There was an error and the gateway could not be removed from bookmarks',
})

const GatewayOverviewHeader = ({ gateway }) => {
  const [deleteGatewayVisible, setDeleteGatewayVisible] = useState(false)

  const dispatch = useDispatch()
  const { ids, name } = gateway
  const { gateway_id } = ids
  const user = useSelector(selectUser)
  const bookmarks = useSelector(selectBookmarksList)
  const addBookmarkLoading = useSelector(state => selectFetchingEntry(state, ADD_BOOKMARK_BASE))
  const deleteBookmarkLoading = useSelector(state =>
    selectFetchingEntry(state, DELETE_BOOKMARK_BASE),
  )
  const mayDeleteGtw = useSelector(state => checkFromState(mayDeleteGateway, state))
  const supportsClaiming = useSelector(selectSelectedGatewayClaimable)

  const isBookmarked = useMemo(
    () => bookmarks.map(b => b.entity_ids?.gateway_ids?.gateway_id).some(b => b === gateway_id),
    [bookmarks, gateway_id],
  )

  const handleAddToBookmark = useCallback(async () => {
    try {
      if (!isBookmarked) {
        await dispatch(attachPromise(addBookmark(user.ids.user_id, { gateway_ids: ids })))
        return
      }
      await dispatch(
        attachPromise(
          deleteBookmark(user.ids.user_id, {
            name: 'gateway',
            id: gateway_id,
          }),
        ),
      )
    } catch (e) {
      toast({
        title: gateway_id,
        message: isBookmarked ? m.removeBookmarkFail : m.addBookmarkFail,
        type: toast.types.ERROR,
      })
    }
  }, [dispatch, gateway_id, ids, isBookmarked, user.ids])

  const handleGlobalConfDownload = useCallback(async () => {
    try {
      const globalConf = await tts.Gateways.getGlobalConf(gateway_id)
      const globalConfDataUri = composeDataUri(JSON.stringify(globalConf, undefined, 2))
      downloadDataUriAsFile(globalConfDataUri, 'global_conf.json')
    } catch (err) {
      toast({
        title: sharedMessages.globalConfFailed,
        message: sharedMessages.globalConfFailedMessage,
        type: toast.types.ERROR,
      })
    }
  }, [gateway_id])

  const handleOpenDeleteGatewayModal = useCallback(() => {
    setDeleteGatewayVisible(true)
  }, [])

  const menuDropdownItems = (
    <>
      <Dropdown.Item title={sharedMessages.downloadGlobalConf} action={handleGlobalConfDownload} />
      {/* <Dropdown.Item title={m.duplicateGateway} action={() => {}} />*/}
      {mayDeleteGtw && (
        <Dropdown.Item
          title={
            supportsClaiming ? sharedMessages.unclaimAndDeleteGateway : sharedMessages.deleteGateway
          }
          action={handleOpenDeleteGatewayModal}
          messageClassName="c-text-error-normal"
        />
      )}
    </>
  )

  return (
    <div className={style.root}>
      <div className="overflow-hidden d-flex flex-column gap-cs-xs">
        <h5 className={style.name}>{name || gateway_id}</h5>
        <span className={style.id}>
          <Message className={style.idPrefix} content={sharedMessages.id} uppercase />
          {gateway_id}
        </span>
      </div>
      <div className="d-inline-flex h-full al-center gap-cs-m flex-wrap">
        <GatewayConnection className="md-lg:d-none" gtwId={gateway_id} />
        <div className={classnames(style.divider, 'md-lg:d-none')} />
        <div className="d-inline-flex al-center gap-cs-xxs">
          <Button
            secondary
            icon={!isBookmarked ? IconStar : IconStarFilled}
            onClick={handleAddToBookmark}
            disabled={
              (!isBookmarked && addBookmarkLoading) || (isBookmarked && deleteBookmarkLoading)
            }
            tooltip={
              isBookmarked ? sharedMessages.removeFromBookmarks : sharedMessages.addToBookmarks
            }
          />
          <Button
            secondary
            icon={IconMenu2}
            noDropdownIcon
            dropdownItems={menuDropdownItems}
            dropdownPosition="below left"
            data-test-id="gateway-overview-menu"
          />
        </div>
        <DeleteEntityHeaderModal
          entity={GATEWAY}
          entityId={gateway_id}
          entityName={name}
          setVisible={setDeleteGatewayVisible}
          visible={deleteGatewayVisible}
          supportsClaiming={supportsClaiming}
        />
      </div>
    </div>
  )
}

GatewayOverviewHeader.propTypes = {
  gateway: PropTypes.gateway.isRequired,
}

export default GatewayOverviewHeader
