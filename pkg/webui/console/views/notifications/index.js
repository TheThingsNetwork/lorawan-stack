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

import React from 'react'
import { useSelector } from 'react-redux'
import { Container } from 'react-grid-system'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import WithRootClass from '@ttn-lw/lib/components/with-root-class'

import NotificationsContainer from '@console/containers/notifications'

import style from '@console/views/app/app.styl'

import { getUnseenNotifications } from '@console/store/actions/notifications'

import { selectUserId } from '@console/store/selectors/logout'

import styles from './notifications.styl'

const NotificationsView = () => {
  const userId = useSelector(selectUserId)

  return (
    <RequireRequest requestAction={getUnseenNotifications(userId)}>
      <WithRootClass className={style.stageBg} id="stage">
        <Container className={styles.mobileContainer}>
          <NotificationsContainer />
        </Container>
      </WithRootClass>
    </RequireRequest>
  )
}

export default NotificationsView
