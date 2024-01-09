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

import React, { useEffect, useRef } from 'react'
import lottie from 'lottie-web'

import PropTypes from '@ttn-lw/lib/prop-types'

const Animation = props => {
  const containerRef = useRef(null)

  const { animationData, lottieConfig, animationRef, ...rest } = props

  useEffect(() => {
    animationRef.current = lottie.loadAnimation({
      container: containerRef.current,
      renderer: 'svg',
      loop: false,
      autoplay: false,
      animationData,
      ...lottieConfig,
    })

    return () => {
      animationRef.current = null
    }
  }, [animationData, animationRef, lottieConfig])

  return <div ref={containerRef} {...rest} />
}

Animation.propTypes = {
  animationData: PropTypes.shape({}).isRequired,
  animationRef: PropTypes.shape({ current: PropTypes.any }).isRequired,
  lottieConfig: PropTypes.shape({}),
}

Animation.defaultProps = {
  lottieConfig: {},
}

export default Animation
