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

import React, { useCallback, useEffect, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Col, Row } from 'react-grid-system'
import classNames from 'classnames'
import { defineMessages } from 'react-intl'

import Pagination from '@ttn-lw/components/pagination'
import Button from '@ttn-lw/components/button'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import useQueryState from '@ttn-lw/lib/hooks/use-query-state'

import { getNotifications } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'
import { selectTotalNotificationsCount } from '@console/store/selectors/notifications'

import NotificationList from './notification-list'
import NotificationContent from './notification-content'

import style from './notifications.styl'

const m = defineMessages({
  seeArchived: 'See archived messages',
  seeAll: 'See all messages',
})

const pageSize = 5
const DEFAULT_PAGE = 1

const pageValidator = page => (!Boolean(page) || page < 0 ? DEFAULT_PAGE : page)

const Notifications = ({ setPage, page }) => {
  const userId = useSelector(selectUserId)
  const totalNotifications = useSelector(selectTotalNotificationsCount)
  const dispatch = useDispatch()
  const [selectedNotification, setSelectedNotification] = useState(undefined)
  const [showArchived, setShowArchived] = useQueryState('archived', 'false')

  const fetchItems = useCallback(
    async filter => {
      const filters =
        showArchived === 'true'
          ? ['NOTIFICATION_STATUS_ARCHIVED']
          : filter ?? ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN']
      await dispatch(
        attachPromise(
          getNotifications(userId, filters, {
            limit: pageSize,
            page,
          }),
        ),
      )
    },
    [dispatch, userId, page, showArchived],
  )

  useEffect(() => {
    fetchItems()
  }, [page, fetchItems])

  const onPageChange = useCallback(
    page => {
      setPage(pageValidator(page))
    },
    [setPage],
  )

  const handleShowArchived = useCallback(async () => {
    setPage(DEFAULT_PAGE)
    setShowArchived(showArchived === 'false' ? 'true' : 'false')
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
          setSelectedNotification={setSelectedNotification}
          selectedNotification={selectedNotification}
          isArchive={showArchived === 'true'}
        />
        <Row direction="column" align="center" className="mt-cs-xxl">
          <Pagination
            pageCount={Math.ceil(totalNotifications / pageSize) || 1}
            onPageChange={onPageChange}
            disableInitialCallback
            pageRangeDisplayed={2}
            forcePage={page}
          />
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
            fetchItems={fetchItems}
            isArchive={showArchived === 'true'}
          />
        )}
      </Col>
    </Row>
  )
}

Notifications.propTypes = {
  page: PropTypes.number.isRequired,
  setPage: PropTypes.func.isRequired,
}

export default Notifications
