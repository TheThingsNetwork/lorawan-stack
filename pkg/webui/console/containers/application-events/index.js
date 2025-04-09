// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useMemo } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import Events from '@console/components/events'

import Require from '@console/lib/components/require'

import PropTypes from '@ttn-lw/lib/prop-types'

import { mayViewApplicationEvents } from '@console/lib/feature-checks'

import {
  clearApplicationEventsStream,
  pauseApplicationEventsStream,
  resumeApplicationEventsStream,
  setApplicationEventsFilter,
} from '@console/store/actions/applications'

import {
  selectApplicationEvents,
  selectApplicationEventsPaused,
  selectApplicationEventsTruncated,
  selectApplicationEventsFilter,
  selectApplicationById,
} from '@console/store/selectors/applications'
import { selectConsolePreferences } from '@console/store/selectors/user-preferences'

const m = defineMessages({
  applicationEventsOf: 'Application events of <strong>{entityName}</strong>',
})

const ApplicationEvents = props => {
  const { appId, widget, darkTheme, framed } = props

  const applicationName = useSelector(state => selectApplicationById(state, appId).name) || appId

  const events = useSelector(state => selectApplicationEvents(state, appId))
  const paused = useSelector(state => selectApplicationEventsPaused(state, appId))
  const truncated = useSelector(state => selectApplicationEventsTruncated(state, appId))
  const filter = useSelector(state => selectApplicationEventsFilter(state, appId))
  const consolePreferences = useSelector(selectConsolePreferences)
  const dark =
    consolePreferences.console_theme === 'CONSOLE_THEME_DARK' ||
    (consolePreferences.console_theme === 'CONSOLE_THEME_SYSTEM' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  const dispatch = useDispatch()

  const onClear = useCallback(() => {
    dispatch(clearApplicationEventsStream(appId))
  }, [appId, dispatch])

  const onPauseToggle = useCallback(
    paused => {
      if (paused) {
        dispatch(resumeApplicationEventsStream(appId))
        return
      }
      dispatch(pauseApplicationEventsStream(appId))
    },
    [appId, dispatch],
  )

  const onFilterChange = useCallback(
    filterId => {
      dispatch(setApplicationEventsFilter(appId, filterId))
    },
    [appId, dispatch],
  )

  const content = useMemo(() => {
    if (widget) {
      return (
        <Events.Widget entityId={appId} events={events} toAllUrl={`/applications/${appId}/data`} />
      )
    }

    return (
      <Events
        entityId={appId}
        events={events}
        paused={paused}
        onClear={onClear}
        truncated={truncated}
        filter={filter}
        onPauseToggle={onPauseToggle}
        onFilterChange={onFilterChange}
        darkTheme={dark ?? darkTheme}
        framed={framed}
        titleMessage={m.applicationEventsOf}
        entityName={applicationName}
      />
    )
  }, [
    appId,
    applicationName,
    darkTheme,
    dark,
    events,
    filter,
    framed,
    onClear,
    onFilterChange,
    onPauseToggle,
    paused,
    truncated,
    widget,
  ])

  return <Require featureCheck={mayViewApplicationEvents}>{content}</Require>
}

ApplicationEvents.propTypes = {
  appId: PropTypes.string.isRequired,
  darkTheme: PropTypes.bool,
  framed: PropTypes.bool,
  widget: PropTypes.bool,
}

ApplicationEvents.defaultProps = {
  darkTheme: false,
  framed: false,
  widget: false,
}

export default ApplicationEvents
