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

import React from 'react'
import { Redirect } from 'react-router-dom'
import bind from 'autobind-decorator'

import Require from './require'

/**
 * `withFeatureRequirement` is a HOC that checks whether the current has the
 * necessary authorization to view the wrapped component. It can be set up to
 * either redirect to another route, to render something different or to not
 * render anything if the requirement is not met.
 * @param {Object} featureCheck - The feature check object containing the right
 * selector as well as the check itself.
 * @param {Object} otherwise - A configuration object determining what should be
 * rendered if the requirement was not met. If not set, nothing will be rendered.
 * @returns {Function} - An instance of the `withFeatureRequirement` HOC.
 */
const withFeatureRequirement = (featureCheck, otherwise) => Component =>
  class WithFeatureRequirement extends React.Component {
    @bind
    alternativeRender() {
      if (typeof otherwise === 'object') {
        const { render, redirect } = otherwise

        if (typeof redirect === 'function') {
          return <Redirect to={redirect(this.props)} />
        } else if (typeof redirect === 'string') {
          return <Redirect to={redirect} />
        } else if (typeof render === 'function') {
          return render()
        }
      }

      return null
    }

    render() {
      return (
        <Require featureCheck={featureCheck} alternativeRender={this.alternativeRender}>
          <Component {...this.props} />
        </Require>
      )
    }
  }

export default withFeatureRequirement
