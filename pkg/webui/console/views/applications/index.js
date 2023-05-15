// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { Routes, Route } from 'react-router-dom'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import NotFoundRoute from '@ttn-lw/lib/components/not-found-route'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import Application from '@console/views/application'
import ApplicationsList from '@console/views/applications-list'
import ApplicationAdd from '@console/views/application-add'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import { pathId as pathIdRegexp } from '@ttn-lw/lib/regexp'

import { mayViewApplications } from '@console/lib/feature-checks'

const Applications = props => {
  const { path } = props.match

  useBreadcrumbs('apps', <Breadcrumb path="/applications" content={sharedMessages.applications} />)

  return (
    <Routes>
      <Route exact path={`${path}`} component={ApplicationsList} />
      <Route exact path={`${path}/add`} component={ApplicationAdd} />
      <Route path={`${path}/:appId${pathIdRegexp}`} component={Application} sensitive />
      <NotFoundRoute />
    </Routes>
  )
}
Applications.propTypes = {
  match: PropTypes.match.isRequired,
}
export default withFeatureRequirement(mayViewApplications, { redirect: '/' })(Applications)
