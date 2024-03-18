// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import Panel from '@ttn-lw/components/panel'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { selectDocumentationUrlConfig } from '@ttn-lw/lib/selectors/env'

import styles from './documentation-dashboard-panel.styl'

const docBaseUrl = selectDocumentationUrlConfig()

const m = defineMessages({
  tts: 'The Things Stack',
  reference: 'Reference',
  gettingStarted: 'Getting started',
})

const DocsPanelLink = ({ path, title, icon }) => (
  <Link to={path} className={styles.docsLink} target="_blank">
    <div className="d-flex al-center gap-cs-xs">
      <Icon icon={icon} /> <Message content={title} />
    </div>
    <Icon icon="arrow-right" />
  </Link>
)

DocsPanelLink.propTypes = {
  icon: PropTypes.string.isRequired,
  path: PropTypes.string.isRequired,
  title: PropTypes.message.isRequired,
}

const DocumentationDashboardPanel = () => (
  <Panel
    title={sharedMessages.documentation}
    path={docBaseUrl}
    icon="book"
    buttonTitle={sharedMessages.documentation}
    divider
    target="_blank"
    className="h-full"
  >
    <DocsPanelLink
      path={`${docBaseUrl}/getting-started/`}
      title={m.gettingStarted}
      icon="users-group"
    />
    <DocsPanelLink path={`${docBaseUrl}/devices/`} title={sharedMessages.devices} icon="device" />
    <DocsPanelLink
      path={`${docBaseUrl}/gateways/`}
      title={sharedMessages.gateways}
      icon="gateway"
    />
    <DocsPanelLink
      path={`${docBaseUrl}/integrations/`}
      title={sharedMessages.integrations}
      icon="arrow-merge"
    />
    <DocsPanelLink path={`${docBaseUrl}/the-things-stack/`} title={m.tts} icon="tts" />
    <DocsPanelLink path={`${docBaseUrl}/reference/`} title={m.reference} icon="book" />
  </Panel>
)

export default DocumentationDashboardPanel
