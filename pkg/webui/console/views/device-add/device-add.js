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
import { Switch, Route, Redirect } from 'react-router-dom'
import { Container, Col, Row } from 'react-grid-system'
import { defineMessages } from 'react-intl'

import PageTitle from '@ttn-lw/components/page-title'
import Tabs from '@ttn-lw/components/tabs'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import PropTypes from '@ttn-lw/lib/prop-types'

import Manual from './manual'
import Repository from './repository'
import messages from './messages'

import style from './device-add.styl'

const m = defineMessages({
  title: 'Register end device',
})

const DeviceAdd = props => {
  const { match } = props
  const { url } = match

  const tabs = React.useMemo(
    () => [
      { title: messages.repositoryTabTitle, link: `${url}/repository`, name: 'repository' },
      { title: messages.manualTabTitle, link: `${url}/manual`, name: 'manual', exact: false },
    ],
    [url],
  )

  return (
    <Container>
      <Row>
        <Col>
          <PageTitle title={m.title} />
          <Tabs className={style.tabs} narrow tabs={tabs} />
        </Col>
      </Row>
      <Switch>
        <Redirect exact from={url} to={`${url}/repository`} />
        <Route path={`${match.url}/repository`} component={Repository} />
        <Route path={`${match.url}/manual`} component={Manual} />
        <NotFoundRoute />
      </Switch>
    </Container>
  )
}

DeviceAdd.propTypes = {
  match: PropTypes.match.isRequired,
}

export default DeviceAdd
