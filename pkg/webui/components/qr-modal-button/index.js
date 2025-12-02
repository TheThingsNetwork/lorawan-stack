// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { defineMessages } from 'react-intl'

import { IconCamera, IconPlus } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'
import ModalButton from '@ttn-lw/components/button/modal-button'
import Button from '@ttn-lw/components/button'
import Input from '@ttn-lw/components/input'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectPluginTTSCloud } from '@ttn-lw/lib/selectors/env'

import DataSheet from '../data-sheet'
import QR from '../qr'

const smUrl = selectPluginTTSCloud().subscription_management_url

const QrScanDoc = (
  <Link.Anchor external secondary href="https://www.thethingsindustries.com/docs/">
    Having trouble?
  </Link.Anchor>
)

const m = defineMessages({
  scanContinue: 'Please scan the QR code to continue. {qrScanDoc}',
  apply: 'Apply',
  fleetToken: 'Fleet owner token',
  addFleet: 'Add to Fleet',
  addToFleetTooltip:
    'You are registering a Managed gateway. If you want to add it to an existing fleet, click here.',
})

const QRModalButton = props => {
  const { message, onApprove, onCancel, onRead, qrData, setQrData, invalidMessage } = props
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

  const handleRead = useCallback(
    val => {
      onRead(val)
    },
    [onRead],
  )

  const modalData = (
    <div style={{ width: '100%' }}>
      {qrData.data ? (
        qrData.valid ? (
          <>
            <DataSheet data={qrData.data} />
            {qrData.gateway.is_managed && (
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
            )}
          </>
        ) : (
          <ErrorMessage content={invalidMessage} />
        )
      ) : (
        <>
          <QR onChange={handleRead} />
          <Message
            content={m.scanContinue}
            values={{ qrScanDoc: QrScanDoc }}
            component="span"
            className="c-text-neutral-light"
          />
        </>
      )}
    </div>
  )

  return (
    <ModalButton
      type="button"
      icon={IconCamera}
      onCancel={onCancel}
      onApprove={onApprove}
      message={message}
      modalData={{
        title: message,
        children: modalData,
        buttonMessage: m.apply,
        approveButtonProps: {
          primary: true,
          disabled: !qrData.valid,
        },
        cancelButtonMessage: qrData.data ? sharedMessages.scanAgain : sharedMessages.cancel,
        cancelButtonProps: qrData.data ? { onClick: onCancel } : {},
        danger: false,
      }}
    />
  )
}

QRModalButton.propTypes = {
  invalidMessage: PropTypes.message.isRequired,
  message: PropTypes.message.isRequired,
  onApprove: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  onRead: PropTypes.func.isRequired,
  qrData: PropTypes.shape({
    valid: PropTypes.bool,
    data: PropTypes.arrayOf(PropTypes.shape()),
    gateway: PropTypes.shape({
      is_managed: PropTypes.bool,
    }),
  }),
  setQrData: PropTypes.func.isRequired,
}

QRModalButton.defaultProps = {
  qrData: undefined,
}

export default QRModalButton
