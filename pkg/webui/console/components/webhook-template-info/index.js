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
  moreInfo: 'More info',
  documentation: 'Documentation',
  templateInformation: 'Template information',
})

const WebhookTemplateInfo = ({ webhookTemplate }) => {
  const { logo_url, name, description, info_url, documentation_url } = webhookTemplate
  const showDivider = Boolean(info_url) && Boolean(documentation_url)
  return (
    <>
      <Message className={style.heading} component="h4" content={m.templateInformation} />
      <div className={style.templateInfo}>
        <div className={style.logo}>
          <img alt={name} src={logo_url} />
        </div>
        <div>
          <h3 className={style.name}>{name}</h3>
          <span className={style.description}>{description}</span>
          <span className={style.info}>
            {info_url && (
              <Link.Anchor target="_blank" href={info_url}>
                <Message content={m.moreInfo} />
              </Link.Anchor>
            )}
            {showDivider && <span className={style.divider}>|</span>}
            {documentation_url && (
              <Link.Anchor target="_blank" href={documentation_url}>
                <Message content={m.documentation} />
              </Link.Anchor>
            )}
          </span>
        </div>
      </div>
    </>
  )
}

WebhookTemplateInfo.propTypes = {
  webhookTemplate: PropTypes.webhookTemplate.isRequired,
}

export default WebhookTemplateInfo
