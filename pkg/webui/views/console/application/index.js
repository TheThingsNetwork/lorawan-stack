// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import { connect } from 'react-redux'

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'

@connect(function (_, props) {
  return { appId: props.match.params.appId }
})
@withBreadcrumb('apps.single', function (props) {
  const { appId } = props
  return (
    <Breadcrumb
      path={`/console/applications/${appId}`}
      icon="application"
      content={appId}
    />
  )
})
export default class Application extends React.Component {

  render () {
    const { appId } = this.props
    return <div><strong>{appId}</strong> application</div>
  }
}
