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

import React, { useEffect, useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import classNames from 'classnames'

import devicePlaceholder from '@assets/misc/placeholder-device.svg'

import Panel from '@ttn-lw/components/panel'
import Icon, {
  IconWorld,
  IconFileAnalytics,
  IconDevice,
  IconAccessPoint,
} from '@ttn-lw/components/icon'
import ButtonGroup from '@ttn-lw/components/button/group'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import TagList from '@console/components/tag-list'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'

import { getBrand, listModels } from '@console/store/actions/device-repository'

import { selectSelectedDevice } from '@console/store/selectors/devices'
import { selectDeviceModelById } from '@console/store/selectors/device-repository'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device-info-panel.styl'

const m = defineMessages({
  endDeviceInfo: 'End device info',
  deviceRepository: 'Device repository',
  deviceWebsite: 'Device website',
  rssi: `{rssi}dBm`,
  snr: `{snr}dB`,
  notInDeviceRepository: 'This end device is currently not part of our Device Repository',
})

const hasDecodedPayload = data => {
  const type = data?.['@type']?.split('.')?.pop()

  return (
    type === 'ApplicationUplink' ||
    type === 'ApplicationUplinkNormalized' ||
    type === 'ApplicationUp'
  )
}

const DeviceInfoPanel = ({ events }) => {
  const [brandName, setBrandName] = useState('')
  const [latestEvent, setLatestEvent] = useState(null)
  const appId = useSelector(selectSelectedApplicationId)
  const device = useSelector(selectSelectedDevice)

  const actualLastEvent = events.find(e => hasDecodedPayload(e.data))
  // Save latestEvent only if it there are 10 seconds between actualLastEvent and lastEvent (throttling).
  if (
    actualLastEvent &&
    (!latestEvent || Date.parse(actualLastEvent.time) - Date.parse(latestEvent.time) > 10000)
  ) {
    setLatestEvent(actualLastEvent)
  }

  const { version_ids = {} } = device
  const hasVersionIds = Object.keys(version_ids).length > 0
  const shortCutLinkPath = hasVersionIds
    ? `https://www.thethingsnetwork.org/device-repository/devices/${version_ids.brand_id}/${version_ids.model_id}/`
    : 'https://www.thethingsnetwork.org/device-repository'

  const model = useSelector(state =>
    selectDeviceModelById(state, version_ids.brand_id, version_ids.model_id),
  )

  const deviceImage = model?.photos?.main || devicePlaceholder

  const dispatch = useDispatch()
  const handleGetBrand = useCallback(async () => {
    const brand = await dispatch(attachPromise(getBrand(appId, version_ids.brand_id, ['name'])))
    setBrandName(brand?.name)
  }, [appId, version_ids.brand_id, dispatch])

  useEffect(() => {
    if (Object.keys(version_ids).length > 0) {
      handleGetBrand()
    }
  }, [handleGetBrand, version_ids])

  return (
    <Panel
      title={m.endDeviceInfo}
      icon={IconDevice}
      shortCutLinkTitle={m.deviceRepository}
      shortCutLinkPath={shortCutLinkPath}
      shortCutLinkTarget="_blank"
      divider
    >
      {Object.keys(version_ids).length > 0 ? (
        <RequireRequest
          requestAction={listModels(appId, version_ids.brand_id, {}, [
            'name',
            'photos',
            'product_url',
            'datasheet_url',
            'sensors',
          ])}
        >
          <div className="d-flex gap-cs-xl">
            <div className={style.deviceImage}>
              <img
                className={classNames({
                  'h-full': !model?.photos?.main,
                  [style.deviceImagePlaceholder]: !model?.photos?.main,
                })}
                src={deviceImage}
                name={model?.name}
              />
            </div>
            <div className="d-flex direction-column j-center gap-cs-m" style={{ lineHeight: 1 }}>
              <Message content={model?.name} className="fw-bold" />
              <Message content={brandName} />
              <div className="d-flex gap-cs-m">
                {latestEvent && (
                  <>
                    <div className="d-inline-flex al-center gap-cs-xxs">
                      <Icon icon={IconAccessPoint} />
                      <Message
                        content={m.snr}
                        values={{
                          snr: latestEvent?.data.uplink_message?.rx_metadata?.[0]?.snr ?? 0,
                        }}
                      />
                    </div>
                    <div className="d-inline-flex al-center gap-cs-xxs">
                      <Icon icon={IconAccessPoint} />
                      <Message
                        content={m.rssi}
                        values={{
                          rssi: latestEvent?.data.uplink_message?.rx_metadata?.[0]?.rssi ?? 0,
                        }}
                      />
                    </div>
                  </>
                )}
                {/* Battery */}
              </div>
              <ButtonGroup align="start">
                <Button.AnchorLink
                  secondary
                  href={model?.product_url}
                  target="_blank"
                  message={m.deviceWebsite}
                  icon={IconWorld}
                />
                <Button.AnchorLink
                  secondary
                  href={model?.datasheet_url}
                  target="_blank"
                  message={sharedMessages.dataSheet}
                  icon={IconFileAnalytics}
                />
              </ButtonGroup>
            </div>
          </div>
          {model?.sensors && <TagList tags={model?.sensors} />}
        </RequireRequest>
      ) : (
        <div className="d-flex gap-cs-xl">
          <div className={style.deviceImage}>
            <img className={style.deviceImagePlaceholder} src={devicePlaceholder} />
          </div>
          <div className="d-flex direction-column j-center gap-cs-m" style={{ lineHeight: 1 }}>
            <span className="fw-bold">{device.name ?? device.ids.device_id}</span>
            <Message content={m.notInDeviceRepository} className="c-text-neutral-light" />
            {latestEvent && (
              <div className="d-flex gap-cs-m">
                <div className="d-inline-flex al-center gap-cs-xxs">
                  <Icon icon={IconAccessPoint} />
                  <Message
                    content={m.snr}
                    values={{
                      snr: latestEvent?.data.uplink_message?.rx_metadata?.[0]?.snr ?? 0,
                    }}
                  />
                </div>
                <div className="d-inline-flex al-center gap-cs-xxs">
                  <Icon icon={IconAccessPoint} />
                  <Message
                    content={m.rssi}
                    values={{
                      rssi: latestEvent?.data.uplink_message?.rx_metadata?.[0]?.rssi ?? 0,
                    }}
                  />
                </div>
                {/* Battery */}
              </div>
            )}
          </div>
        </div>
      )}
    </Panel>
  )
}

DeviceInfoPanel.propTypes = {
  events: PropTypes.events.isRequired,
}

export default DeviceInfoPanel
