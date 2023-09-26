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

import React, { useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Col, Row } from 'react-grid-system'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import useQueryState from '@ttn-lw/lib/hooks/use-query-state'

import { getNotifications } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'

import NotificationList from './notification-list'
import NotificationContent from './notification-content'

import style from './notifications.styl'

const m = defineMessages({
  seeArchived: 'See archived messages',
  seeAll: 'See all messages',
})

const pageSize = 5
const DEFAULT_PAGE = 1

const Notifications = () => {
  const userId = useSelector(selectUserId)
  const dispatch = useDispatch()
  const [selectedNotification, setSelectedNotification] = useState(undefined)
  const [hasNextPage, setHasNextPage] = useState(true)
  const [isNextPageLoading, setIsNextPageLoading] = useState(false)
  const [items, setItems] = useState([])
  const [page, setPage] = useState(DEFAULT_PAGE)
  const [showArchived, setShowArchived] = useQueryState('archived', 'false')

  const loadNextPage = useCallback(
    async filter => {
      setIsNextPageLoading(true)
      const filters =
        showArchived === 'true'
          ? ['NOTIFICATION_STATUS_ARCHIVED']
          : typeof filter === 'string'
          ? filter
          : ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN']
      const newItems = await dispatch(
        attachPromise(
          getNotifications(userId, filters, {
            limit: pageSize,
            page,
          }),
        ),
      )
      setPage(page => page + 1)
      setItems(items => [...items, ...newItems.notifications])
      setHasNextPage(items.length < newItems.totalCount)
      setIsNextPageLoading(false)
    },
    [dispatch, userId, page, setPage, showArchived, items],
  )

  const handleShowArchived = useCallback(async () => {
    setPage(DEFAULT_PAGE)
    setShowArchived(showArchived === 'false' ? 'true' : 'false')
    setItems([])
    setHasNextPage(true)
  }, [setShowArchived, showArchived, setPage])

  return (
    <Row className={classNames(style.notificationsContainer, 'm-0')}>
      <Col
        md={4}
        className={classNames(style.notificationList, 'mt-cs-l', 'mb-cs-l', {
          [style.notificationSelected]: selectedNotification,
        })}
      >
        <NotificationList
          hasNextPage={hasNextPage}
          isNextPageLoading={isNextPageLoading}
          loadNextPage={loadNextPage}
          items={items}
          setSelectedNotification={setSelectedNotification}
          selectedNotification={selectedNotification}
          isArchive={showArchived === 'true'}
        />
        <Row direction="column" align="center" className="mt-cs-xxl">
          <Button
            onClick={handleShowArchived}
            naked
            message={showArchived === 'true' ? m.seeAll : m.seeArchived}
            className={style.notificationListChangeButton}
          />
        </Row>
      </Col>
      <Col
        md={8}
        className={classNames(style.notificationContent, 'mt-cs-l', 'mb-cs-l', 'p-0', {
          [style.notificationSelected]: selectedNotification,
        })}
      >
        {selectedNotification && (
          <NotificationContent
            setSelectedNotification={setSelectedNotification}
            selectedNotification={selectedNotification}
            fetchItems={loadNextPage}
            isArchive={showArchived === 'true'}
          />
        )}
      </Col>
    </Row>
  )
}

export default Notifications
