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

import Pagination from '@ttn-lw/components/pagination'

import Message from '@ttn-lw/lib/components/message'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getNotifications } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'
import { selectTotalNotificationsCount } from '@console/store/selectors/notifications'

import NotificationList from './notification-list'
import NotificationContent from './notification-content'

import style from './notifications.styl'

const pageSize = 6

const DEFAULT_PAGE = 1

const pageValidator = page => (!Boolean(page) || page < 0 ? DEFAULT_PAGE : page)

const NotificationsContainer = ({ setPage, page }) => {
  const userId = useSelector(selectUserId)
  const totalNotifications = useSelector(selectTotalNotificationsCount)
  const dispatch = useDispatch()
  const [archiving, setArchiving] = useState(false)
  const [selectedNotification, setSelectedNotification] = useState(undefined)

  const fetchItems = useCallback(
    async filter => {
      const filters = filter ?? ['NOTIFICATION_STATUS_UNSEEN', 'NOTIFICATION_STATUS_SEEN']
      await dispatch(
        attachPromise(
          getNotifications(userId, filters, {
            limit: pageSize,
            page,
          }),
        ),
      )
    },
    [dispatch, userId, page],
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

  return (
    <Row className={classNames(style.notificationsContainer, 'm-0')}>
      <Col md={4.5} className={classNames(style.notificationList, 'mt-cs-l', 'mb-cs-l')}>
        <NotificationList
          setSelectedNotification={setSelectedNotification}
          selectedNotification={selectedNotification}
          setArchiving={setArchiving}
          fetchItems={fetchItems}
        />
        <Row direction="column" align="center">
          <Pagination
            pageCount={Math.ceil(totalNotifications / pageSize) || 1}
            onPageChange={onPageChange}
            disableInitialCallback
            pageRangeDisplayed={2}
            forcePage={page}
          />
          <Message content="See archived messages" />
        </Row>
      </Col>
      <Col md={7.5} className={classNames(style.notificationContent, 'mt-cs-l', 'mb-cs-l', 'p-0')}>
        {selectedNotification && !archiving && (
          <NotificationContent
            selectedNotification={selectedNotification}
            setArchiving={setArchiving}
            fetchItems={fetchItems}
          />
        )}
      </Col>
    </Row>
  )
}

NotificationsContainer.propTypes = {
  page: PropTypes.number.isRequired,
  setPage: PropTypes.func.isRequired,
}

export default NotificationsContainer
