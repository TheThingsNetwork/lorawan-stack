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

import Notification from '@ttn-lw/components/notification'
import Link from '@ttn-lw/components/link'

import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  hintTitle: 'Your end device will be added soon!',
  hintMessage:
    "We're sorry, but your device is not yet part of The LoRaWAN Device Repository. You can use <ManualLink>manual device registration</ManualLink>, using the information your end device manufacturer provided e.g. in the product's data sheet. Please also refer to our documentation on <GuideLink>Adding Devices</GuideLink>.",
})

const OtherHint = React.memo(props => {
  const { manualLinkPath, manualGuideDocsPath } = props

  return (
    <Notification
      info
      small
      title={m.hintTitle}
      content={m.hintMessage}
      messageValues={{
        ManualLink: msg => (
          <Link secondary key="manual-link" to={manualLinkPath}>
            {msg}
          </Link>
        ),
        GuideLink: msg => (
          <Link.DocLink secondary key="manual-guide-link" path={manualGuideDocsPath}>
            {msg}
          </Link.DocLink>
        ),
      }}
    />
  )
})

OtherHint.propTypes = {
  manualGuideDocsPath: PropTypes.string,
  manualLinkPath: PropTypes.string.isRequired,
}

OtherHint.defaultProps = {
  manualGuideDocsPath: undefined,
}

export default OtherHint
