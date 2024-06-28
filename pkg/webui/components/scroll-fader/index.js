// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useEffect, useRef } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'
import combineRefs from '@ttn-lw/lib/combine-refs'

import style from './scroll-fader.styl'

// ScrollFader is a component that fades out the content of a container when it
// is scrolled. It is used for scrollable elements that need some visual
// indication that they are scrollable, but do not have a scrollbar.
// The indication only shows when the content is scrolled.

const ScrollFader = React.forwardRef(
  ({ children, className, fadeHeight, light, faderHeight, topFaderOffset }, ref) => {
    const internalRef = useRef()
    const combinedRef = combineRefs([ref, internalRef])

    const handleScroll = useCallback(() => {
      const container = internalRef.current
      const { scrollTop, scrollHeight, clientHeight } = container
      const scrollable = scrollHeight - clientHeight
      const scrollGradientTop = container.querySelector(`.${style.scrollGradientTop}`)
      const scrollGradientBottom = container.querySelector(`.${style.scrollGradientBottom}`)

      if (scrollGradientTop) {
        const opacity = scrollTop < fadeHeight ? scrollTop / fadeHeight : 1
        scrollGradientTop.style.opacity = opacity
      }

      if (scrollGradientBottom) {
        const scrollEnd = scrollable - fadeHeight
        const opacity = scrollTop < scrollEnd ? 1 : (scrollable - scrollTop) / fadeHeight
        scrollGradientBottom.style.opacity = opacity
      }
    }, [fadeHeight])

    useEffect(() => {
      const container = internalRef.current
      if (!container) return

      const mutationObserver = new MutationObserver(() => {
        handleScroll()
      })

      // Run the calculation whenever the children change.
      mutationObserver.observe(container, { attributes: false, childList: true, subtree: false })

      handleScroll() // Call once on mount if needed
      container.addEventListener('wheel', handleScroll)
      window.addEventListener('resize', handleScroll)

      return () => {
        // Cleanup observer and event listeners
        mutationObserver.disconnect()
        container.removeEventListener('wheel', handleScroll)
        window.removeEventListener('resize', handleScroll)
      }
    }, [handleScroll])

    return (
      <div
        ref={combinedRef}
        onScroll={handleScroll}
        className={className}
        style={{
          position: 'relative',
        }}
      >
        <div
          className={classnames(style.scrollGradientTop, { [style.scrollGradientTopLight]: light })}
          style={{
            height: faderHeight,
            marginBottom: `-${faderHeight}`,
            top: topFaderOffset,
          }}
        />
        {children}
        <div
          className={classnames(style.scrollGradientBottom, {
            [style.scrollGradientBottomLight]: light,
          })}
          style={{
            height: faderHeight,
            marginTop: `-${faderHeight}`,
          }}
        />
      </div>
    )
  },
)

ScrollFader.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  fadeHeight: PropTypes.number,
  faderHeight: PropTypes.string,
  light: PropTypes.bool,
  topFaderOffset: PropTypes.string,
}

ScrollFader.defaultProps = {
  className: undefined,
  fadeHeight: 40,
  faderHeight: '1rem',
  light: false,
  topFaderOffset: '0',
}

export default ScrollFader
