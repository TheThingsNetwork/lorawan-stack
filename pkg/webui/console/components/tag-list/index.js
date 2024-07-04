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

import React, { useState, useEffect } from 'react'
import { upperFirst } from 'lodash'

import Icon, { IconAccessPoint } from '@ttn-lw/components/icon'
import Tooltip from '@ttn-lw/components/tooltip'

import PropTypes from '@ttn-lw/lib/prop-types'

const TagList = ({ tags }) => {
  const containerRef = React.useRef()
  const [visibleTags, setVisibleTags] = useState([])
  const [hiddenTagsCount, setHiddenTagsCount] = useState(0)

  // Function to calculate text width
  const getTextWidth = (text, font) => {
    const canvas = document.createElement('canvas')
    const context = canvas.getContext('2d')
    context.font = font
    return context.measureText(text).width
  }

  useEffect(() => {
    const calculateVisibleTags = () => {
      if (containerRef.current) {
        const font = window.getComputedStyle(containerRef.current).font
        let currentWidth = 0
        let visibleCount = 0

        // Add padding and margin to each tag
        const tagPadding = 21
        const tagMargin = 10.5

        const containerWidth =
          containerRef.current.offsetWidth -
          21 -
          (getTextWidth(tags[tags.length - 1], font) + tagPadding)
        for (let i = 0; i < tags.length; i++) {
          const tagWidth = getTextWidth(tags[i], font) + tagPadding + tagMargin
          if (currentWidth + tagWidth <= containerWidth) {
            currentWidth += tagWidth
            visibleCount++
          } else {
            break
          }
        }

        // Check if we need to add the "more" tag
        if (visibleCount < tags.length) {
          const moreTagText = `+${tags.length - visibleCount}`
          const moreTagWidth = getTextWidth(moreTagText, font) + tagPadding + tagMargin

          // If adding the "more" tag pushes us over the limit, remove one more visible tag
          if (currentWidth + moreTagWidth > containerWidth) {
            visibleCount--
          }
        }

        setVisibleTags(tags.slice(0, visibleCount))
        setHiddenTagsCount(Math.max(0, tags.length - visibleCount))
      }
    }

    calculateVisibleTags()
    window.addEventListener('resize', calculateVisibleTags)

    return () => {
      window.removeEventListener('resize', calculateVisibleTags)
    }
  }, [tags])

  return (
    <div
      ref={containerRef}
      className="d-flex j-strat gap-cs-s mt-cs-xl overflow-hidden"
      style={{ flexWrap: 'nowrap' }}
    >
      {visibleTags.map(tag => (
        <span
          key={tag}
          className="d-flex j-center al-center gap-cs-xxs p-sides-cs-s p-vert-cs-xxs br-xl c-bg-neutral-light"
          style={{ textWrap: 'nowrap' }}
        >
          <Icon icon={IconAccessPoint} />
          {upperFirst(tag)}
        </span>
      ))}
      {hiddenTagsCount > 0 && (
        <Tooltip
          content={tags
            .slice(tags.length - hiddenTagsCount)
            .map(tag => upperFirst(tag))
            .join(', ')}
        >
          <div className="d-flex j-center al-center gap-cs-xxs p-sides-cs-s p-vert-cs-xxs br-xl c-bg-neutral-light">
            +{hiddenTagsCount}
          </div>
        </Tooltip>
      )}
    </div>
  )
}

TagList.propTypes = {
  tags: PropTypes.arrayOf(PropTypes.string).isRequired,
}

export default TagList
