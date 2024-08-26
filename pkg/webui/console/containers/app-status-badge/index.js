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

import React, { useMemo } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import classnames from 'classnames'

import OnlineStatus from '@ttn-lw/constants/online-status'

import {
  IconRefreshAlert,
  IconNetworkOff,
  IconCloudOff,
  IconTrendingDown,
} from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'
import { useAlertBanner } from '@ttn-lw/components/alert-banner/context'
import StatusLabel from '@ttn-lw/components/status-label'

import StatusIndicator from '@console/containers/app-status-badge/status-indicator'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import {
  selectNetworkStatusIndicatorStore,
  selectOnlineStatus,
  selectUpcomingMaintenancesStore,
} from '@ttn-lw/lib/store/selectors/status'
import { selectPageStatusBaseUrlConfig } from '@ttn-lw/lib/selectors/env'

import alertStyle from '../../../components/alert-banner/alert-banner.styl'
import statusLabelStyle from '../../../components/status-label/status-label.styl'

import style from './app-status-badge.styl'

const m = defineMessages({
  upcomingMaintenance: 'Upcoming <link>maintenance</link>',
  degradedPerformance: '<link>Degraded performance</link>',
  maintenanceInProgress: '<link>Maintenance</link> in progress',
  networkOfflineTitle: 'Network offline',
  networkOfflineSubtitle: 'The console lost internet connection',
  degradedPerformanceTitle: 'Degraded performance',
  degradedPerformanceSubtitle: 'Check the <link>status page</link> for more info',
  connectionIssuesSubtitle: 'The console is having trouble connecting to the internet',
  maintenance: 'Maintenance',
  maintenanceScheduledSubtitle: '<link>Maintenance</link> scheduled in less than 1 hour',
})

// NOTE: The network status is fetched during initialization in the init logic.

const AppStatusBadge = ({ className }) => {
  const onlineStatus = useSelector(selectOnlineStatus)
  const statusPageUrl = selectPageStatusBaseUrlConfig()
  const statusIndicator = useSelector(selectNetworkStatusIndicatorStore)
  const upcomingMaintenances = useSelector(selectUpcomingMaintenancesStore)
  const { showBanner } = useAlertBanner()

  const status = useMemo(() => {
    if (onlineStatus === OnlineStatus.OFFLINE) {
      return {
        type: 'error',
        label: sharedMessages.offline,
        icon: IconNetworkOff,
        alertTitle: m.networkOfflineTitle,
        alertSubtitle: m.networkOfflineSubtitle,
      }
    } else if (onlineStatus === OnlineStatus.CHECKING) {
      return {
        type: 'warning',
        label: sharedMessages.connectionIssues,
        icon: IconCloudOff,
        alertTitle: sharedMessages.connectionIssues,
        alertSubtitle: m.connectionIssuesSubtitle,
      }
    } else if (statusIndicator !== StatusIndicator.NONE) {
      return {
        type: 'warning',
        label: m.degradedPerformance,
        icon: IconTrendingDown,
        alertTitle: m.degradedPerformanceTitle,
        alertSubtitle: m.degradedPerformanceSubtitle,
      }
    } else if (upcomingMaintenances.length) {
      const currentTime = new Date()
      const scheduledFor = new Date(upcomingMaintenances[0].scheduled_for)
      const diff = scheduledFor - currentTime
      const hoursUntilMaintenance = diff / 1000 / 60 / 60

      if (hoursUntilMaintenance > 0 && hoursUntilMaintenance < 1) {
        return {
          type: 'info',
          label: m.upcomingMaintenance,
          icon: IconRefreshAlert,
          alertTitle: m.maintenance,
          alertSubtitle: m.maintenanceScheduledSubtitle,
        }
      }

      const scheduledUntil = new Date(upcomingMaintenances[0].scheduled_until)

      if (currentTime >= scheduledFor && currentTime <= scheduledUntil) {
        return {
          type: 'info',
          label: m.maintenanceInProgress,
          icon: IconRefreshAlert,
          alertTitle: m.maintenance,
          alertSubtitle: m.maintenanceInProgress,
        }
      }
    }

    return null
  }, [onlineStatus, statusIndicator, upcomingMaintenances])

  const handleBadgeClick = () => {
    showBanner({
      type: status.type,
      title: status.alertTitle,
      subtitle: status.alertSubtitle,
      subtitleValues: {
        link: link => (
          <Link.Anchor className={alertStyle.link} href={statusPageUrl} target="_blank">
            <span>{link}</span>
          </Link.Anchor>
        ),
      },
    })
  }

  if (!status) {
    return null
  }

  return (
    <StatusLabel
      type={status.type}
      icon={status.icon}
      content={status.label}
      className={classnames(style.appStatusBadge, className)}
      contentValues={{
        link: link => (
          <Link className={statusLabelStyle.link} to={statusPageUrl} target="_blank">
            <span>{link}</span>
          </Link>
        ),
      }}
      onClick={handleBadgeClick}
    />
  )
}

AppStatusBadge.propTypes = {
  className: PropTypes.string,
}

AppStatusBadge.defaultProps = {
  className: undefined,
}
export default AppStatusBadge
