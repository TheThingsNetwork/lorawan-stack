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

import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import style from '@account/views/front/front.styl'

import { selectApplicationSiteName, selectApplicationSiteTitle } from '@ttn-lw/lib/selectors/env'
import errorMessages from '@ttn-lw/lib/errors/error-messages'
import statusCodeMessages from '@ttn-lw/lib/errors/status-code-messages'
import sharedMessages from '@ttn-lw/lib/shared-messages'

const siteName = selectApplicationSiteName()
const siteTitle = selectApplicationSiteTitle()

const FrontNotFound = () => (
  <React.Fragment>
    <div className={style.form}>
      <IntlHelmet title={statusCodeMessages['404']} />
      <h1 className={style.title}>
        {siteName}
        <br />
        <Message content={statusCodeMessages['404']} component="strong" />
      </h1>
      <hr className={style.hRule} />
      <Message
        content={errorMessages.genericNotFound}
        component="p"
        className={style.errorDescription}
      />
      <Button.Link
        to="/login"
        icon="keyboard_arrow_left"
        message={{ ...sharedMessages.backTo, values: { siteTitle } }}
      />
    </div>
  </React.Fragment>
)

export default FrontNotFound
