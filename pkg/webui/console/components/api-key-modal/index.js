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

import Icon, { IconCheck } from '@ttn-lw/components/icon'
import PortalledModal from '@ttn-lw/components/modal/portalled'
import SafeInspector from '@ttn-lw/components/safe-inspector'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './api-key-modal.styl'

const m = defineMessages({
  title: 'Please copy newly created API key',
  subtitle: "You won't be able to view the key afterward",
  buttonMessage: 'I have copied the key',
  grantedRights: 'Granted rights',
  description: `Your API key has been created successfully.
Note: After closing this window, the value of the key secret will not be accessible anymore.
Make sure to copy and store it in a safe place now.`,
})

const ApiKeyModal = props => {
  const { visible, secret, rights, ...rest } = props

  if (!visible) {
    return null
  }

  return (
    <PortalledModal
      visible={visible}
      {...rest}
      title={m.title}
      subtitle={m.subtitle}
      approval={false}
      buttonMessage={m.buttonMessage}
    >
      <div className={style.left}>
        <Message component="h4" content={m.grantedRights} />
        <ul>
          {rights.map(right => (
            <li key={right}>
              <Icon icon={IconCheck} className={style.icon} />
              <Message className={style.rightName} content={{ id: `enum:${right}` }} firstToUpper />
            </li>
          ))}
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
  rights: PropTypes.rights,
  secret: PropTypes.string,
  visible: PropTypes.bool.isRequired,
}

ApiKeyModal.defaultProps = {
  rights: undefined,
  secret: undefined,
}

export default ApiKeyModal
