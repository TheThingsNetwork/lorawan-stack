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
import { Navigate, useSearchParams } from 'react-router-dom'
import { defineMessages } from 'react-intl'

import { IconChevronLeft } from '@ttn-lw/components/icon'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Button from '@ttn-lw/components/button'
import PageTitle from '@ttn-lw/components/page-title'

import Message from '@ttn-lw/lib/components/message'

import { selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const m = defineMessages({
  codeDescription: 'Your authorization code is:',
})

const siteTitle = selectApplicationSiteTitle()

const Code = () => {
  const [searchParams] = useSearchParams()
  const code = searchParams.get('code')

  if (!code) {
    return <Navigate to="/" />
  }

  return (
    <div className="container container--lg grid">
      <div className="item-12 lg-xl:item-6 xl:item-4">
        <div className="d-flex flex-column al-start gap-cs-m">
          <PageTitle title={sharedMessages.authorizationCode} />
          <Message content={m.codeDescription} component="label" className="d-block" />
          <SafeInspector data={code} initiallyVisible hideable={false} isBytes={false} />
          <Button.Link
            to="/"
            icon={IconChevronLeft}
            message={{ ...sharedMessages.backTo, values: { siteTitle } }}
            secondary
          />
        </div>
      </div>
    </div>
  )
}

export default Code
