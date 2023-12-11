// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import style from './webhook-template-info.styl'

const m = defineMessages({
  about: 'About {name}',
  setupWebhook: 'Setup webhook for {name}',
  editWebhook: 'Edit webhook for {name}',
})

const WebhookTemplateInfo = ({ webhookTemplate, update }) => {
  const { logo_url, name, description, info_url, documentation_url } = webhookTemplate
  const showDivider = Boolean(info_url) && Boolean(documentation_url)
  return (
    <>
      <div className={style.templateInfo}>
        <div className={style.logo}>
          <img alt={name} src={logo_url} />
        </div>
        <div className={style.descriptionBox}>
          <Message
            component="h3"
            content={update ? m.editWebhook : m.setupWebhook}
            values={{ name }}
            className="m-0"
          />
          <span className="m-0">{description}</span>
          <span className={style.info}>
            {info_url && (
              <Link.Anchor primary href={info_url} external>
                <Message content={m.about} values={{ name }} />
              </Link.Anchor>
            )}
            {showDivider && <span className="mr-cs-xxs ml-cs-xxs">|</span>}
            {documentation_url && (
              <Link.Anchor primary href={documentation_url} external>
                <Message content={m.documentation} />
              </Link.Anchor>
            )}
          </span>
        </div>
      </div>
      <hr className="mb-ls-s" />
    </>
  )
}

WebhookTemplateInfo.propTypes = {
  update: PropTypes.bool,
  webhookTemplate: PropTypes.webhookTemplate.isRequired,
}

WebhookTemplateInfo.defaultProps = {
  update: false,
}

export default WebhookTemplateInfo
