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

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import classnames from 'classnames'

import Tooltip from '@ttn-lw/components/tooltip'

import PropTypes from '@ttn-lw/lib/prop-types'

import Tag from '..'

import style from './group.styl'

const measureWidth = element => {
  if (!element) {
    return 0
  }
  return element.clientWidth
}

const LEFT_TAG_WIDTH = 40
const TAG_SPACE_WIDTH = 3

const TagGroup = ({ className, tagMaxWidth, tags }) => {
  const [left, setLeft] = useState(0)
  const element = useRef(null)

  const handleWindowResize = useCallback(() => {
    const containerWidth = measureWidth(element.current)
    const totalTagCount = tags.length
    const possibleFitCount = Math.floor(containerWidth / tagMaxWidth) || 1

    const leftTagWidth = totalTagCount !== possibleFitCount ? LEFT_TAG_WIDTH : 0
    const spaceWidth = possibleFitCount > 1 ? possibleFitCount * TAG_SPACE_WIDTH : 0

    const finalAvailableWidth = containerWidth - leftTagWidth - spaceWidth
    const finalLeft = Math.floor(finalAvailableWidth / tagMaxWidth) || 1

    setLeft(totalTagCount - finalLeft)
  }, [tagMaxWidth, tags.length])

  useEffect(() => {
    window.addEventListener('resize', handleWindowResize)
    handleWindowResize()

    return () => {
      window.removeEventListener('resize', handleWindowResize)
    }
  }, [handleWindowResize, tags])

  const ts = useMemo(() => tags.slice(0, tags.length - left), [left, tags])
  const leftGroup = <div className={style.leftGroup}>{tags.slice(tags.length - left)}</div>

  return (
    <div ref={element} className={classnames(className, style.group)}>
      {ts}
      {left > 0 && (
        <Tooltip content={leftGroup}>
          <Tag content={`+${left}`} />
        </Tooltip>
      )}
    </div>
  )
}

TagGroup.propTypes = {
  className: PropTypes.string,
  tagMaxWidth: PropTypes.number.isRequired,
  tags: PropTypes.arrayOf(PropTypes.shape(Tag.PropTypes)).isRequired,
}

TagGroup.defaultProps = {
  className: undefined,
}

export default TagGroup
