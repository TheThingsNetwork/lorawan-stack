// Copyright © 2025 The Things Industries B.V.

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'

import { IconPlus } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'
import Button from '@ttn-lw/components/button'
import Input from '@ttn-lw/components/input'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectPluginTTSCloud } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

const smUrl = selectPluginTTSCloud().subscription_management_url

const m = defineMessages({
  addFleet: 'Add to Fleet',
  fleetToken: 'Fleet owner token',
  addToFleetTooltip:
    'You are registering a Managed gateway. If you want to add it to an existing fleet, click here.',
})

const FleetsScan = ({ qrData, setQrData }) => {
  const [isAddToFleet, setIsAddToFleet] = React.useState(undefined)
  const [fleetOwnerToken, setFleetOwnerToken] = React.useState('')

  const handleAddToFleet = useCallback(() => {
    setIsAddToFleet(true)
  }, [])

  const handleRemoveFleet = useCallback(() => {
    setIsAddToFleet(false)
    setFleetOwnerToken('')
    setQrData({
      ...qrData,
      gateway: { ...qrData.gateway, _fleet_owner_token: undefined },
    })
  }, [qrData, setQrData])

  const handleFleetTokenChange = useCallback(
    value => {
      setFleetOwnerToken(value)
      setQrData({
        ...qrData,
        gateway: { ...qrData.gateway, _fleet_owner_token: value },
      })
    },
    [qrData, setQrData],
  )

  return qrData.gateway.is_managed ? (
    <>
      {isAddToFleet ? (
        <div className="w-full mt-cs-xs">
          <div className="d-flex j-between">
            <Message content={m.fleetToken} className="c-text-neutral-semilight" />
            <Button
              message={sharedMessages.remove}
              onClick={handleRemoveFleet}
              className="c-text-neutral-light fw-bold hover-underline"
            />
          </div>
          <Message
            content={sharedMessages.fleetTokenInfo}
            values={{
              Link: val => (
                <Link.Anchor secondary href={`${smUrl}/dashboard`} external>
                  {val}
                </Link.Anchor>
              ),
            }}
            className="mb-cs-xs mt-cs-xxs c-text-neutral-light fs-s"
            component="div"
          />
          <Input
            type="code"
            sensitive
            className="w-full"
            inputWidth="full"
            value={fleetOwnerToken}
            onChange={handleFleetTokenChange}
          />
          <Message
            content={sharedMessages.fleetInfo}
            className="mt-cs-xs c-text-neutral-light fs-s"
            component="div"
          />
        </div>
      ) : (
        <Button
          tertiary
          message={m.addFleet}
          icon={IconPlus}
          onClick={handleAddToFleet}
          className="mt-cs-xs"
          tooltip={m.addToFleetTooltip}
          tooltipPlacement="bottom"
        />
      )}
    </>
  ) : null
}

FleetsScan.propTypes = {
  qrData: PropTypes.shape({
    gateway: PropTypes.shape({
      is_managed: PropTypes.bool,
    }),
  }).isRequired,
  setQrData: PropTypes.func.isRequired,
}

export default FleetsScan
