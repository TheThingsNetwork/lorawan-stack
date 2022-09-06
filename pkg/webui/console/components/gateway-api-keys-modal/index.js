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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './gateway-api-key-modal.styl'

const m = defineMessages({
  modalTitle: 'Gateway API keys',
  buttonMessage: 'Go to gateway',
  grantedRights: 'Granted rights',
  description: `<b>Your Gateway has been created successfully.</b>{lineBreak}Below are the API keys.{lineBreak}
  Note: After proceeding to the gateway, the API keys will not be accessible for download anymore.
  Make sure to copy and store them in a safe place now.`,
  lnsKey: 'LNS key:',
  downloadButton: 'Download',
  cupsKey: 'CUPS key:',
})

const GatewayApiKeysModal = ({ modalVisible, lnsKey, cupsKey, downloadLns, downloadCups }) => (
  <PortalledModal
    visible={modalVisible}
    title={m.modalTitle}
    approval={false}
    buttonMessage={m.buttonMessage}
  >
    <div className={style.div}>
      <Message
        content={m.description}
        values={{ b: str => <b>{str}</b>, lineBreak: <br /> }}
        component="p"
      />
      {lnsKey && (
        <div className={style.row}>
          <Message component="h4" content={m.lnsKey} />
          <Button
            type="button"
            message={m.downloadButton}
            onClick={downloadLns}
            icon="file_download"
          />
        </div>
      )}
      {cupsKey && (
        <div className={style.row}>
          <Message component="h4" content={m.cupsKey} />
          <Button
            type="button"
            message={m.downloadButton}
            onClick={downloadCups}
            icon="file_download"
          />
        </div>
      )}
    </div>
  </PortalledModal>
)

GatewayApiKeysModal.propTypes = {
  cupsKey: PropTypes.string,
  downloadCups: PropTypes.func.isRequired,
  downloadLns: PropTypes.func.isRequired,
  lnsKey: PropTypes.string,
  modalVisible: PropTypes.bool,
}

GatewayApiKeysModal.defaultProps = {
  cupsKey: undefined,
  lnsKey: undefined,
  modalVisible: false,
}

export default GatewayApiKeysModal
