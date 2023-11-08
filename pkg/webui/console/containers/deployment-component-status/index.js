// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
import { Col, Row } from 'react-grid-system'

import ServerIcon from '@assets/auxiliary-icons/server.svg'

import Link from '@ttn-lw/components/link'
import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import {
  selectPageStatusBaseUrlConfig,
  selectStackConfig,
  selectDocumentationUrlConfig,
} from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './component-status.styl'

const m = defineMessages({
  availableComponents: 'Available components',
  versionInfo: 'Deployment',
  statusPage: 'Go to status page',
  seeChangelog: 'See changelog',
})

const statusPageBaseUrl = selectPageStatusBaseUrlConfig()
const documentationBaseUrl = selectDocumentationUrlConfig()
const stackConfig = selectStackConfig()
const componentMap = {
  is: sharedMessages.componentIs,
  gs: sharedMessages.componentGs,
  ns: sharedMessages.componentNs,
  as: sharedMessages.componentAs,
  js: sharedMessages.componentJs,
  edtc: sharedMessages.componentEdtc,
  qrg: sharedMessages.componentQrg,
  gcs: sharedMessages.componentGcs,
  dcs: sharedMessages.componentDcs,
}

const DeploymentComponentStatus = () => (
  <Row className="m-vert-ls-l m:mb-ls-xxs">
    <Col sm={4} className="d-flex direction-column">
      <Message content={m.versionInfo} component="h3" className="panel-title" />
      <span className={style.versionValue}>TTS v{process.env.VERSION}</span>
      <pre className="mt-0 fs-s mb-cs-s">{process.env.REVISION}</pre>
      <Link.Anchor href={statusPageBaseUrl} external secondary>
        <Message content={m.statusPage} />
      </Link.Anchor>
      <Link.Anchor
        href={documentationBaseUrl ? `${documentationBaseUrl}/whats-new/` : undefined}
        external
        secondary
      >
        <Message content={m.seeChangelog} />
      </Link.Anchor>
    </Col>
    <Col sm={8} className="d-flex direction-column">
      <Message className="panel-title" content={m.availableComponents} component="h3" />
      <div className="d-flex flex-wrap mt-cs-m">
        {Object.keys(stackConfig).map(componentKey => {
          if (componentKey === 'language') {
            return null
          }
          const component = stackConfig[componentKey]
          const name = componentMap[componentKey]
          const host = component.enabled ? new URL(component.base_url).host : undefined
          return (
            <ComponentCard key={componentKey} name={name} host={host} enabled={component.enabled} />
          )
        })}
      </div>
    </Col>
  </Row>
)

const ComponentCard = ({ name, enabled, host }) => (
  <div className={style.componentCard}>
    <img src={ServerIcon} className={style.componentCardIcon} />
    <div className={style.componentCardDesc}>
      <div className={style.componentCardName}>
        <Status label={name} status={enabled ? 'good' : 'unknown'} flipped />
      </div>
      <span className={style.componentCardHost} title={host}>
        {enabled ? host : <Message content={sharedMessages.disabled} />}
      </span>
    </div>
  </div>
)

ComponentCard.propTypes = {
  enabled: PropTypes.bool.isRequired,
  host: PropTypes.string.isRequired,
  name: PropTypes.message.isRequired,
}

export default DeploymentComponentStatus
