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

import { APPLICATION } from '@console/constants/entities'

import Icon, {
  IconBroadcast,
  IconMenu2,
  IconStar,
  IconStarFilled,
  IconCpu,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Status from '@ttn-lw/components/status'
import Dropdown from '@ttn-lw/components/dropdown'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import LastSeen from '@console/components/last-seen'

import DeleteEntityHeaderModal from '@console/containers/delete-entity-header-modal'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'

import {
  checkFromState,
  mayDeleteApplication,
  mayViewApplicationDevices,
} from '@console/lib/feature-checks'

import {
  ADD_BOOKMARK_BASE,
  addBookmark,
  DELETE_BOOKMARK_BASE,
  deleteBookmark,
} from '@console/store/actions/user-preferences'
import { getApplicationDeviceCount } from '@console/store/actions/applications'
import { getPubsubsList } from '@console/store/actions/pubsubs'
import { getWebhooksList } from '@console/store/actions/webhooks'

import { selectUser } from '@console/store/selectors/user'
import { selectBookmarksList } from '@console/store/selectors/user-preferences'
import {
  selectApplicationDerivedLastSeen,
  selectApplicationDeviceCount,
  selectSelectedApplication,
} from '@console/store/selectors/applications'
import { selectWebhooksTotalCount } from '@console/store/selectors/webhooks'
import { selectPubsubsTotalCount } from '@console/store/selectors/pubsubs'

import style from './application-overview-header.styl'

const m = defineMessages({
  addBookmarkFail: 'There was an error and the application could not be bookmarked',
  removeBookmarkFail: 'There was an error and the application could not be removed from bookmarks',
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
  noActivityTooltip:
    'The network has not recently registered any activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
  deviceCount: '{devices} End devices',
})

const ApplicationOverviewHeader = () => {
  const [deleteApplicationModalVisible, setDeleteApplicationModalVisible] = useState(false)

  const dispatch = useDispatch()
  const { name, ids } = useSelector(selectSelectedApplication)
  const { application_id } = ids
  const user = useSelector(selectUser)
  const bookmarks = useSelector(selectBookmarksList)
  const addBookmarkLoading = useSelector(state => selectFetchingEntry(state, ADD_BOOKMARK_BASE))
  const deleteBookmarkLoading = useSelector(state =>
    selectFetchingEntry(state, DELETE_BOOKMARK_BASE),
  )
  const webhooksCount = useSelector(selectWebhooksTotalCount)
  const pubsubsCount = useSelector(selectPubsubsTotalCount)
  const hasIntegrations = webhooksCount > 0 || pubsubsCount > 0
  const mayViewDevices = useSelector(state => checkFromState(mayViewApplicationDevices, state))
  const devicesTotalCount = useSelector(state =>
    selectApplicationDeviceCount(state, application_id),
  )
  const lastSeen = useSelector(state => selectApplicationDerivedLastSeen(state, application_id))

  const showLastSeen = Boolean(lastSeen)

  const isBookmarked = useMemo(
    () =>
      bookmarks
        .map(b => b.entity_ids?.application_ids?.application_id)
        .some(b => b === application_id),
    [bookmarks, application_id],
  )

  const handleAddToBookmark = useCallback(async () => {
    try {
      if (!isBookmarked) {
        await dispatch(attachPromise(addBookmark(user.ids.user_id, { application_ids: ids })))
        return
      }
      await dispatch(
        attachPromise(
          deleteBookmark(user.ids.user_id, {
            name: 'application',
            id: application_id,
          }),
        ),
      )
    } catch (e) {
      toast({
        title: application_id,
        message: isBookmarked ? m.removeBookmarkFail : m.addBookmarkFail,
        type: toast.types.ERROR,
      })
    }
  }, [application_id, dispatch, ids, isBookmarked, user.ids.user_id])

  const recentActivity = useMemo(() => {
    let node
    if (showLastSeen) {
      node = (
        <DocTooltip
          interactive
          docPath="/getting-started/console/troubleshooting"
          content={<Message content={m.lastSeenAvailableTooltip} />}
        >
          <LastSeen lastSeen={lastSeen} className="c-text-neutral-semilight" />
        </DocTooltip>
      )
    } else {
      node = (
        <DocTooltip
          content={<Message content={m.noActivityTooltip} />}
          docPath="/getting-started/console/troubleshooting"
        >
          <Status
            status="mediocre"
            label={sharedMessages.noRecentActivity}
            className={style.status}
          />
        </DocTooltip>
      )
    }
    return (
      <Link
        to={`/applications/${application_id}/data`}
        className="d-inline-flex al-center gap-cs-xxs md-lg:d-none"
      >
        <Icon icon={IconBroadcast} small className="c-text-neutral-semilight" />
        {node}
      </Link>
    )
  }, [lastSeen, showLastSeen, application_id])

  const handleOpenDeleteApplicationModal = useCallback(() => {
    setDeleteApplicationModalVisible(true)
  }, [])

  const menuDropdownItems = (
    <>
      {
        <Require featureCheck={mayDeleteApplication}>
          <Dropdown.Item
            title={sharedMessages.deleteApp}
            action={handleOpenDeleteApplicationModal}
            messageClassName="c-text-error-normal"
          />
        </Require>
      }
    </>
  )

  return (
    <div className={style.root}>
      <div className="overflow-hidden d-flex flex-column gap-cs-xs">
        <h5 className={style.name}>{name || application_id}</h5>
        <span className={style.id}>
          <Message className={style.idPrefix} content={sharedMessages.id} uppercase />
          {application_id}
        </span>
      </div>
      <div className="d-inline-flex h-full al-center gap-cs-m flex-wrap">
        {recentActivity}
        {mayViewDevices && (
          <RequireRequest requestAction={getApplicationDeviceCount(application_id)}>
            <Link
              to={`/applications/${application_id}/devices`}
              className="d-inline-flex al-center gap-cs-xxs md-lg:d-none"
            >
              <Icon icon={IconCpu} small className="c-text-neutral-semilight" />
              <Message
                content={m.deviceCount}
                className="c-text-neutral-semilight"
                values={{ devices: devicesTotalCount }}
              />
            </Link>
          </RequireRequest>
        )}
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
          />
        </div>
        <RequireRequest
          requestAction={[getPubsubsList(application_id), getWebhooksList(application_id)]}
        >
          <DeleteEntityHeaderModal
            entity={APPLICATION}
            entityId={application_id}
            entityName={name}
            setVisible={setDeleteApplicationModalVisible}
            visible={deleteApplicationModalVisible}
            isPristine={hasIntegrations}
          />
        </RequireRequest>
      </div>
    </div>
  )
}

export default ApplicationOverviewHeader
