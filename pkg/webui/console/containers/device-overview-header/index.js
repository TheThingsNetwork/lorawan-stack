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

import React, { useCallback, useMemo } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages, FormattedNumber } from 'react-intl'
import classnames from 'classnames'

import tts from '@console/api/tts'

import Icon, {
  IconMenu2,
  IconStar,
  IconStarFilled,
  IconArrowsSort,
  IconBroadcast,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import toast from '@ttn-lw/components/toast'
import Tooltip from '@ttn-lw/components/tooltip'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Status from '@ttn-lw/components/status'
import Dropdown from '@ttn-lw/components/dropdown'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import LastSeen from '@console/components/last-seen'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectFetchingEntry } from '@ttn-lw/lib/store/selectors/fetching'
import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'
import { selectNsConfig } from '@ttn-lw/lib/selectors/env'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import {
  ADD_BOOKMARK_BASE,
  addBookmark,
  DELETE_BOOKMARK_BASE,
  deleteBookmark,
} from '@console/store/actions/user-preferences'

import { selectUser } from '@console/store/selectors/logout'
import { selectBookmarksList } from '@console/store/selectors/user-preferences'
import {
  selectDeviceDerivedAppDownlinkFrameCount,
  selectDeviceDerivedNwkDownlinkFrameCount,
  selectDeviceDerivedUplinkFrameCount,
  selectDeviceLastSeen,
  selectSelectedCombinedDeviceId,
} from '@console/store/selectors/devices'

import style from './device-overview-header.styl'

const m = defineMessages({
  addBookmarkFail: 'There was an error and the end device could not be bookmarked',
  removeBookmarkFail: 'There was an error and the end device could not be removed from bookmarks',
  uplinkDownlinkTooltip:
    'The number of sent uplinks and received downlinks of this end device since the last frame counter reset.{break}`App`: frame counter for application downlinks (FPort >=1). `Nwk`: frame counter for network downlinks (FPort = 0)',
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered the last activity of this end device. This is determined from sent uplinks, confirmed downlinks or (re)join requests.{lineBreak}The last activity was received at {lastActivityInfo}',
  noActivityTooltip:
    'The network has not registered any activity from this end device yet. This could mean that your end device has not sent any messages yet or only messages that cannot be handled by the network, e.g. due to a mismatch of EUIs or frequencies.',
  downloadMacData: 'Download MAC data',
  macStateError: 'There was an error and MAC state could not be included in the MAC data.',
})

const nsHost = getHostFromUrl(selectNsConfig().base_url)
const nsEnabled = selectNsConfig().enabled

const DeviceOverviewHeader = ({ device }) => {
  const dispatch = useDispatch()
  const { ids, name } = device
  const { device_id } = ids
  const [appId, devId] = useSelector(selectSelectedCombinedDeviceId).split('/')
  const uplinkFrameCount = useSelector(state =>
    selectDeviceDerivedUplinkFrameCount(state, appId, devId),
  )
  const downlinkAppFrameCount = useSelector(state =>
    selectDeviceDerivedAppDownlinkFrameCount(state, appId, devId),
  )
  const downlinkNwkFrameCount = useSelector(state =>
    selectDeviceDerivedNwkDownlinkFrameCount(state, appId, devId),
  )
  const lastSeen = useSelector(state => selectDeviceLastSeen(state, appId, devId))
  const showLastSeen = Boolean(lastSeen)
  const showUplinkCount = typeof uplinkFrameCount === 'number'
  const showAppDownlinkCount = typeof downlinkAppFrameCount === 'number'
  const showNwkDownlinkCount = typeof downlinkNwkFrameCount === 'number'

  const notAvailableElem = <Message content={sharedMessages.notAvailable} />
  const uplinkValue = showUplinkCount ? (
    <FormattedNumber value={uplinkFrameCount} />
  ) : (
    notAvailableElem
  )
  const downlinkValue =
    showAppDownlinkCount && showNwkDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkAppFrameCount} /> {'(App) , '}
        <FormattedNumber value={downlinkNwkFrameCount} /> {'(Nwk)'}
      </>
    ) : showAppDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkAppFrameCount} /> {'(App)'}
      </>
    ) : showNwkDownlinkCount ? (
      <>
        <FormattedNumber value={downlinkNwkFrameCount} /> {'(Nwk)'}
      </>
    ) : (
      notAvailableElem
    )
  const lastActivityInfo = lastSeen ? <DateTime value={lastSeen} noTitle /> : lastSeen
  const lineBreak = <br />
  const user = useSelector(selectUser)
  const bookmarks = useSelector(selectBookmarksList)
  const addBookmarkLoading = useSelector(state => selectFetchingEntry(state, ADD_BOOKMARK_BASE))
  const deleteBookmarkLoading = useSelector(state =>
    selectFetchingEntry(state, DELETE_BOOKMARK_BASE),
  )

  const isBookmarked = useMemo(
    () => bookmarks.map(b => b.entity_ids?.device_ids?.device_id).some(b => b === device_id),
    [bookmarks, device_id],
  )

  const handleAddToBookmark = useCallback(async () => {
    try {
      if (!isBookmarked) {
        await dispatch(attachPromise(addBookmark(user.ids.user_id, { device_ids: ids })))
        return
      }
      await dispatch(
        attachPromise(
          deleteBookmark(user.ids.user_id, {
            name: 'device',
            id: device_id,
          }),
        ),
      )
    } catch (e) {
      toast({
        title: device_id,
        message: isBookmarked ? m.removeBookmarkFail : m.addBookmarkFail,
        type: toast.types.ERROR,
      })
    }
  }, [dispatch, device_id, ids, isBookmarked, user.ids])

  const onExport = useCallback(async () => {
    const { ids, mac_settings, session, network_server_address } = device

    let result
    if (session && nsEnabled && getHostFromUrl(network_server_address) === nsHost) {
      try {
        result = await tts.Applications.Devices.getById(appId, ids.device_id, ['mac_state'])

        if (!('mac_state' in result)) {
          toast({
            title: m.downloadMacData,
            message: m.macStateError,
            type: toast.types.ERROR,
          })
        }
      } catch {
        toast({
          title: m.downloadMacData,
          message: m.macStateError,
          type: toast.types.ERROR,
        })
      }
    }

    const toExport = { mac_state: result?.mac_state, mac_settings }
    const toExportData = composeDataUri(JSON.stringify(toExport, undefined, 2))
    downloadDataUriAsFile(toExportData, `${ids.device_id}_mac_data_${Date.now()}.json`)
  }, [appId, device])

  const menuDropdownItems = (
    <>
      <Dropdown.Item title={m.downloadMacData} action={onExport} />
    </>
  )

  return (
    <div className={style.root}>
      <div>
        <h5 className={style.name}>{name || device_id}</h5>
        <span className={style.id}>
          <Message className={style.idPrefix} content={sharedMessages.id} uppercase />
          {device_id}
        </span>
      </div>
      <div className="d-inline-flex h-full al-center gap-cs-m flex-wrap">
        <div className="d-flex al-center gap-cs-xxs sm:d-none">
          {showLastSeen ? (
            <DocTooltip
              docPath="/reference/last-activity"
              content={
                <Message
                  content={m.lastSeenAvailableTooltip}
                  values={{ lineBreak, lastActivityInfo }}
                />
              }
            >
              <div className="d-inline-flex al-center gap-cs-xxs">
                <Icon icon={IconBroadcast} small className="c-text-neutral-semilight" />
                <LastSeen lastSeen={lastSeen} className="c-text-neutral-semilight" />
              </div>
            </DocTooltip>
          ) : (
            <DocTooltip
              docPath="/devices/troubleshooting/#my-device-wont-join-what-do-i-do"
              docTitle={sharedMessages.troubleshooting}
              content={<Message content={m.noActivityTooltip} />}
            >
              <div className="d-inline-flex al-center gap-cs-xxs">
                <Icon icon={IconBroadcast} small className="c-text-neutral-semilight" />
                <Status status="mediocre" label={sharedMessages.noActivityYet} />
              </div>
            </DocTooltip>
          )}
        </div>
        <div className="d-flex al-center gap-cs-xxs sm:d-none">
          <Tooltip
            content={
              <Message
                content={m.uplinkDownlinkTooltip}
                values={{ break: <br /> }}
                convertBackticks
              />
            }
          >
            <div className="d-flex al-center gap-cs-xxs">
              <Icon small className="c-text-neutral-semilight" icon={IconArrowsSort} />
              <Message
                component="span"
                content={sharedMessages.upAndDown}
                className="c-text-neutral-semilight"
                values={{
                  up: uplinkValue,
                  down: downlinkValue,
                }}
              />
            </div>
          </Tooltip>
        </div>
        <div className={classnames(style.divider, 'sm:d-none')} />
        <div className="d-inline-flex al-center gap-cs-xxs">
          <Button
            secondary
            icon={!isBookmarked ? IconStar : IconStarFilled}
            onClick={handleAddToBookmark}
            disabled={
              (!isBookmarked && addBookmarkLoading) || (isBookmarked && deleteBookmarkLoading)
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
      </div>
    </div>
  )
}

DeviceOverviewHeader.propTypes = {
  device: PropTypes.device.isRequired,
}

export default DeviceOverviewHeader
