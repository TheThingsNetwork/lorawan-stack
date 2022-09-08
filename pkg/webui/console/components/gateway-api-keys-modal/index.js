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

import React from 'react'
import { defineMessages } from 'react-intl'

import PortalledModal from '@ttn-lw/components/modal/portalled'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './gateway-api-key-modal.styl'

const m = defineMessages({
  modalTitle: 'Download gateway API keys',
  buttonMessage: 'I have downloaded the keys',
  description:
    'Note: After closing this window, these API keys will not be accessible for download anymore. Please make sure to download and store them now.',
  downloadLns: 'Download LNS key',
  downloadCups: 'Download CUPS key',
})

const GatewayApiKeysModal = ({
  modalVisible,
  modalApprove,
  lnsKey,
  cupsKey,
  downloadLns,
  downloadCups,
}) => (
  <PortalledModal
    className={style.gatewayApiKeyModal}
    visible={modalVisible}
    title={m.modalTitle}
    approval={false}
    onComplete={modalApprove}
    buttonMessage={m.buttonMessage}
  >
    <ButtonGroup>
      {lnsKey && (
        <Button type="button" message={m.downloadLns} onClick={downloadLns} icon="file_download" />
      )}
      {cupsKey && (
        <Button
          type="button"
          message={m.downloadCups}
          onClick={downloadCups}
          icon="file_download"
        />
      )}
    </ButtonGroup>
    <Message content={m.description} component="p" />
  </PortalledModal>
)

GatewayApiKeysModal.propTypes = {
  cupsKey: PropTypes.string,
  downloadCups: PropTypes.func.isRequired,
  downloadLns: PropTypes.func.isRequired,
  lnsKey: PropTypes.string,
  modalApprove: PropTypes.func.isRequired,
  modalVisible: PropTypes.bool,
}

GatewayApiKeysModal.defaultProps = {
  cupsKey: undefined,
  lnsKey: undefined,
  modalVisible: false,
}

export default GatewayApiKeysModal
