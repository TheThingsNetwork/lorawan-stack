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

import { IconCamera } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'
import ModalButton from '@ttn-lw/components/button/modal-button'

import Message from '@ttn-lw/lib/components/message'
import ErrorMessage from '@ttn-lw/lib/components/error-message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import DataSheet from '../data-sheet'
import QR from '../qr'

const QrScanDoc = (
  <Link.Anchor external secondary href="https://www.thethingsindustries.com/docs/">
    Having trouble?
  </Link.Anchor>
)

const m = defineMessages({
  scanContinue: 'Please scan the QR code to continue. {qrScanDoc}',
  apply: 'Apply',
})

const QRModalButton = props => {
  const { message, onApprove, onCancel, onRead, qrData, invalidMessage } = props

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
          <DataSheet data={qrData.data} />
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
  }),
}

QRModalButton.defaultProps = {
  qrData: undefined,
}

export default QRModalButton
