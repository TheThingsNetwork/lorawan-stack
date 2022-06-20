// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './hints.styl'

const m = defineMessages({
  hintMessage:
    'Cannot find your exact end device? <SupportLink>Get help here</SupportLink> and <ManualLink>try manual device registration</ManualLink>.',
  hintNoSupportMessage:
    'Cannot find your exact end device? <ManualLink>Try manual device registration</ManualLink>.',
})

const ProgressHint = React.memo(props => {
  const { supportLink, manualLinkPath } = props

  return (
    <Message
      className={style.progressMessage}
      content={Boolean(supportLink) ? m.hintMessage : m.hintNoSupportMessage}
      values={{
        SupportLink: msg => (
          <Link.Anchor secondary key="support-link" href={supportLink} target="_blank">
            {msg}
          </Link.Anchor>
        ),
        ManualLink: msg => (
          <Link secondary key="manual-link" to={manualLinkPath}>
            {msg}
          </Link>
        ),
      }}
    />
  )
})

ProgressHint.propTypes = {
  manualLinkPath: PropTypes.string.isRequired,
  supportLink: PropTypes.string,
}

ProgressHint.defaultProps = {
  supportLink: undefined,
}

export default ProgressHint
