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
import lottie from 'lottie-web'
import bind from 'autobind-decorator'

import PropTypes from '../../prop-types'

@bind
export default class Animation extends Component {
  static propTypes = {
    animationData: PropTypes.object.isRequired,
    lottieConfig: PropTypes.object,
  }

  constructor(props) {
    super(props)

    this.containerRef = React.createRef()
    this.instance = null
  }

  componentDidMount() {
    const { lottieConfig } = this.props

    this.instance = lottie.loadAnimation({
      container: this.containerRef.current,
      renderer: 'svg',
      loop: false,
      autoplay: false,
      animationData: this.props.animationData,
      ...lottieConfig,
    })
  }

  render() {
    const { lottieConfig, animationData, ...rest } = this.props

    return (
      <div
        ref={this.containerRef}
        onMouseEnter={this.handleOnMouseEnter}
        onMouseLeave={this.handleOnMouseLeave}
        {...rest}
      />
    )
  }
}
