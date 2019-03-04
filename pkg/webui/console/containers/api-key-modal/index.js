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

import React from 'react'
import { defineMessages } from 'react-intl'

import PortalledModal from '../../../components/modal/portalled'
import SafeInspector from '../../../components/safe-inspector'
import Icon from '../../../components/icon'
import Message from '../../../lib/components/message'
import sharedMessages from '../../../lib/shared-messages'
import PropTypes from '../../../lib/prop-types'

import style from './api-key-modal.styl'

const m = defineMessages({
  title: 'Please copy newly created API Key',
  subtitle: "You won't be able to view the key afterward",
  buttonMessage: 'I have copied the key',
  grantedRights: 'Granted Rights',
  description: `Your API Key has been created successfully.
Note: After closing this window, the value of the key secret will not be accessible anymore.
Make sure to copy and store it in a safe place now.`,
})

const ApiKeyModal = function (props) {
  const { visible, secret, rights, ...rest } = props

  if (!visible) {
    return null
  }

  return (
    <PortalledModal
      visible={visible}
      modal={{
        ...rest,
        title: m.title,
        subtitle: m.subtitle,
        approval: false,
        buttonMessage: m.buttonMessage,
      }}
    >
      <div className={style.left}>
        <Message component="h4" content={m.grantedRights} />
        <ul>
          {rights.map(right => (
            <li key={right}>
              <Icon icon="check" className={style.icon} />
              <Message className={style.rightName} content={{ id: `enum:${right}` }} />
            </li>
          )
          )}
        </ul>
      </div>
      <div className={style.right}>
        <Message className={style.description} component="p" content={m.description} />
        <Message component="h3" content={sharedMessages.apiKey} />
        <SafeInspector
          className={style.secretInspector}
          data={secret}
          isBytes={false}
          disableResize
        />
      </div>
    </PortalledModal>
  )
}

ApiKeyModal.propTypes = {
  visible: PropTypes.bool.isRequired,
  secret: PropTypes.string,
  rights: PropTypes.arrayOf(PropTypes.string),
}

export default ApiKeyModal
