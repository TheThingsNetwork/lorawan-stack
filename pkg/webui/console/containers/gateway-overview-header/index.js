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

import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import Icon, {
  IconCalendarMonth,
  IconMenu2,
  IconStar,
  IconStarFilled,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import GatewayConnection from '@console/containers/gateway-connection'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { addBookmark, deleteBookmark } from '@console/store/actions/user-preferences'

import { selectUser } from '@console/store/selectors/logout'

import style from './gateway-overview-header.styl'

const m = defineMessages({
  addBookmarkFail: 'There was an error and the gateway could not be bookmarked',
  removeBookmarkFail: 'There was an error and the gateway could not be removed from bookmarks',
})

const GatewayOverviewHeader = ({ gateway }) => {
  const dispatch = useDispatch()
  const { ids, name, created_at } = gateway
  const { gateway_id } = ids
  const user = useSelector(selectUser)

  const isBookmarked = false

  const handleAddToBookmark = useCallback(async () => {
    try {
      if (!isBookmarked) {
        await dispatch(attachPromise(addBookmark(user.ids, ids)))
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

  return (
    <div className={style.root}>
      <div>
        <h5 className={style.name}>{name || gateway_id}</h5>
        <span className={style.id}>
          <Message className={style.idPrefix} content={sharedMessages.id} uppercase />
          {gateway_id}
        </span>
      </div>
      <div className="d-inline-flex h-full al-center gap-cs-m flex-wrap">
        <GatewayConnection gtwId={gateway_id} />
        <div className="d-flex al-center gap-cs-xxs">
          <Icon small className="c-text-neutral-semilight" icon={IconCalendarMonth} />
          <LastSeen
            displayStatus={false}
            message={sharedMessages.created}
            lastSeen={created_at}
            className="c-text-neutral-semilight"
          />
        </div>
        <div className={style.divider} />
        <div className="d-inline-flex al-center gap-cs-xxs">
          <Button
            secondary
            icon={!isBookmarked ? IconStar : IconStarFilled}
            onClick={handleAddToBookmark}
          />
          <Button secondary icon={IconMenu2} />
        </div>
      </div>
    </div>
  )
}

GatewayOverviewHeader.propTypes = {
  gateway: PropTypes.gateway.isRequired,
}

export default GatewayOverviewHeader
