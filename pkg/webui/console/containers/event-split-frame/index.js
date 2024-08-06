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

import React, { useContext, useCallback, useRef, useEffect } from 'react'
import DOM from 'react-dom'

import PropTypes from '@ttn-lw/lib/prop-types'

import EventSplitFrameContext from './context'

import style from './event-split-frame.styl'

const EventSplitFrameInner = ({ children }) => {
  const { isOpen, height, isActive, setHeight, setIsMounted } = useContext(EventSplitFrameContext)
  const ref = useRef()

  useEffect(() => {
    setIsMounted(true)
    return () => setIsMounted(false)
  }, [setIsMounted])

  // Handle the dragging of the handler to resize the frame.
  const handleDragStart = useCallback(
    e => {
      e.preventDefault()

      const startY = e.clientY
      const startHeight = height

      const handleDragMove = e => {
        const newHeight = startHeight + (startY - e.clientY)
        setHeight(Math.max(3 * 14, newHeight))
      }

      const handleDragEnd = () => {
        window.removeEventListener('mousemove', handleDragMove)
        window.removeEventListener('mouseup', handleDragEnd)
      }

      window.addEventListener('mousemove', handleDragMove)
      window.addEventListener('mouseup', handleDragEnd)
    },
    [height, setHeight],
  )

  if (!isActive) {
    return null
  }

  return (
    <div className={style.container} style={{ height }} ref={ref}>
      {isOpen && (
        <>
          <div className={style.header} onMouseDown={handleDragStart}>
            <div className={style.handle} />
          </div>
          <div className={style.content}>{children}</div>
        </>
      )}
    </div>
  )
}

EventSplitFrameInner.propTypes = {
  children: PropTypes.node.isRequired,
}

const EventSplitFrame = props =>
  DOM.createPortal(<EventSplitFrameInner {...props} />, document.getElementById('split-frame'))

export default EventSplitFrame
