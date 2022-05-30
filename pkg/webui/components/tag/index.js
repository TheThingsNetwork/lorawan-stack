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

import React, { useEffect, useRef, useState } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'
import combineRefs from '@ttn-lw/lib/combine-refs'

import Tooltip from '../tooltip'

import style from './tag.styl'

const Tag = React.forwardRef((props, ref) => {
  const { content, className } = props
  const [enableTooltip, setEnableTooltip] = useState(false)
  const widthRef = useRef(ref)

  useEffect(() => {
    if (!widthRef || !widthRef.current) {
      return
    }

    const elem = widthRef.current

    if (elem.offsetWidth < elem.scrollWidth) {
      setEnableTooltip(true)
    } else {
      setEnableTooltip(false)
    }
  }, [widthRef])

  const tag = (
    <div ref={combineRefs([ref, widthRef])} className={classnames(className, style.tag)}>
      {content}
    </div>
  )

  if (enableTooltip) {
    return <Tooltip content={content}>{tag}</Tooltip>
  }

  return tag
})

Tag.propTypes = {
  className: PropTypes.string,
  content: PropTypes.string.isRequired,
}

Tag.defaultProps = {
  className: undefined,
}
export default Tag
