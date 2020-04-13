// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Redirect } from 'react-router'

import ApplicationWebhookAddForm from '@console/views/application-integrations-webhook-add-form'

import PropTypes from '@ttn-lw/lib/prop-types'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectWebhookTemplates } from '@console/store/selectors/webhook-templates'

@connect(state => ({
  appId: selectSelectedApplicationId(state),
  hasTemplates: selectWebhookTemplates(state).length !== 0,
}))
export default class ApplicationWebhookAdd extends Component {
  static propTypes = {
    hasTemplates: PropTypes.bool.isRequired,
    match: PropTypes.match.isRequired,
  }
  render() {
    const {
      match,
      match: { url: path },
      hasTemplates,
    } = this.props

    // Forward to the template chooser, when templates have been configured.
    if (hasTemplates) {
      return <Redirect to={`${path}/template`} />
    }

    return <ApplicationWebhookAddForm match={match} />
  }
}
