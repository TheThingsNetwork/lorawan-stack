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

import React, { useEffect, useRef } from 'react'
import DOM from 'react-dom'

import PropTypes from '@ttn-lw/lib/prop-types'
import combineRefs from '@ttn-lw/lib/combine-refs'

// This component create a portal to render children into a different part of the DOM.
const Portal = ({ children, visible, positionReference }) => {
  const [buttonPosition, setButtonPosition] = React.useState()
  const attachedRef = useRef()
  const buttonRef = combineRefs([positionReference.current, attachedRef])

  useEffect(() => {
    if (positionReference && visible) {
      const rect = positionReference.current.getBoundingClientRect()
      setButtonPosition(rect)
    }
  }, [positionReference, visible])

  if (!buttonPosition) {
    return null
  }

  const recreatedComponent = (
    <>
      {/* Position component in correct part of the screen */}
      <div
        style={{
          position: 'relative',
          top: buttonPosition ? buttonPosition.top + window.scrollY : 0,
          left: buttonPosition ? buttonPosition.left + window.scrollX : 0,
          minWidth: '100%',
          minHeight: '100%',
        }}
        ref={buttonRef}
      >
        {/* Simulate button to allow for the dropdown positioning logic to work correctly */}
        <div
          style={{
            width: buttonPosition ? buttonPosition.width : 0,
            height: buttonPosition ? buttonPosition.height : 0,
          }}
        />
        {children}
      </div>
    </>
  )

  return DOM.createPortal(
    visible && recreatedComponent,
    document.getElementById('dropdown-container'),
  )
}

Portal.propTypes = {
  children: PropTypes.node.isRequired,
  positionReference: PropTypes.shape({
    current: PropTypes.shape({
      getBoundingClientRect: PropTypes.func,
    }),
  }).isRequired,
  visible: PropTypes.bool.isRequired,
}

export default Portal
