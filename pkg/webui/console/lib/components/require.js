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
import bind from 'autobind-decorator'
import { Redirect } from 'react-router-dom'

import PropTypes from '../../../lib/prop-types'

@connect((state, props) => {
  const { featureCheck, condition } = props
  const rights = featureCheck && featureCheck.rightsSelector(state)
  return {
    condition: condition || (Boolean(featureCheck) && featureCheck.check(rights)),
  }
})
export default class Require extends Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    condition: PropTypes.bool.isRequired,
    otherwise: PropTypes.shape({
      redirect: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
      render: PropTypes.func,
    }),
  }
  static defaultProps = {
    otherwise: undefined,
  }

  @bind
  alternativeRender() {
    const { otherwise } = this.props
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
    const { condition, children } = this.props

    if (!condition) {
      return this.alternativeRender()
    }

    return children
  }
}
