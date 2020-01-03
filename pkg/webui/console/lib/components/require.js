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

import { Component } from 'react'
import { connect } from 'react-redux'

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
    alternativeRender: PropTypes.func,
    children: PropTypes.node.isRequired,
    condition: PropTypes.bool.isRequired,
  }
  static defaultProps = {
    alternativeRender: undefined,
  }
  render() {
    const { condition, children, alternativeRender } = this.props

    if (!condition) {
      if (typeof alternativeRender === 'function') {
        return alternativeRender()
      }
      return null
    }

    return children
  }
}
