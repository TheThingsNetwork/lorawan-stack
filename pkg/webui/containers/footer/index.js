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
import { useSelector } from 'react-redux'

import FooterComponent from '@ttn-lw/components/footer'

import { selectOnlineStatus } from '@ttn-lw/lib/store/selectors/status'
import {
  selectDocumentationUrlConfig,
  selectSupportLinkConfig,
  selectPageStatusBaseUrlConfig,
} from '@ttn-lw/lib/selectors/env'

const supportLink = selectSupportLinkConfig()
const documentationBaseUrl = selectDocumentationUrlConfig()
const statusPageBaseUrl = selectPageStatusBaseUrlConfig()

const Footer = props => {
  const onlineStatus = useSelector(selectOnlineStatus)

  return (
    <FooterComponent
      {...props}
      onlineStatus={onlineStatus}
      supportLink={supportLink}
      documentationLink={documentationBaseUrl}
      statusPageLink={statusPageBaseUrl}
    />
  )
}

const { onlineStatus: onlineStatusPropType, ...propTypes } = FooterComponent.propTypes

Footer.propTypes = propTypes

const { onlineStatus: onlineStatusDefaultProp, ...defaultProps } = FooterComponent.defaultProps

Footer.defaultProps = defaultProps

export default Footer
