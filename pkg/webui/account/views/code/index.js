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
import Query from 'query-string'
import { Redirect } from 'react-router-dom'
import { defineMessages } from 'react-intl'

import SafeInspector from '@ttn-lw/components/safe-inspector'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import { selectApplicationSiteName, selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

const m = defineMessages({
  code: 'Authorization code',
  codeDescription: 'Your authorization code is:',
  backToAccount: 'Back to {siteTitle}',
})

const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()

const Code = ({ location }) => {
  const { query } = Query.parseUrl(location.search)

  if (!query.code) {
    return <Redirect to="/" />
  }

  return (
    <>
      <div className={style.form}>
        <IntlHelmet title={m.createANewAccount} />
        <h1 className={style.title}>
          {siteName}
          <br />
          <Message content={m.code} component="strong" />
        </h1>
        <hr className={style.hRule} />
        <Message content={m.codeDescription} component="label" className={style.codeDescription} />
        <SafeInspector
          data={query.code}
          initiallyVisible
          hideable={false}
          isBytes={false}
          className={style.code}
        />
        <Button.Link
          to="/"
          icon="keyboard_arrow_left"
          message={{ ...m.backToAccount, values: { siteTitle } }}
        />
      </div>
    </>
  )
}

Code.propTypes = {
  location: PropTypes.location.isRequired,
}

export default Code
