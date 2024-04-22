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

import React, { useEffect } from 'react'
import DOM from 'react-dom'

import PropTypes from '@ttn-lw/lib/prop-types'

// This component create a portal to render children into a different part of the DOM.
const Portal = ({ children, visible, positionReference }) => {
  const [buttonPosition, setButtonPosition] = React.useState()
  useEffect(() => {
    if (positionReference.current && visible) {
      const rect = positionReference.current.getBoundingClientRect()
      setButtonPosition(rect)
      const elementLeft = rect.left + window.scrollX
      const elementTop = rect.top + window.scrollY
      const element = document.getElementById('modal-container')
      element.style.position = 'absolute'
      element.style.left = `${elementLeft}px`
      element.style.top = `${elementTop}px`
      element.style.minWidth = `${rect.width + elementLeft}px`
      element.style.minHeight = `${rect.height}px`
    }
  }, [positionReference, visible])

  const recreatedComponent = (
    <>
      <div
        className="pos-absolute"
        style={{
          top: 0,
          left: 0,
          visibility: 'hidden',
          width: buttonPosition ? buttonPosition.width : 0,
          height: buttonPosition ? buttonPosition.height : 0,
        }}
      />
      {children}
    </>
  )

  return DOM.createPortal(visible && recreatedComponent, document.getElementById('modal-container'))
}

Portal.propTypes = {
  children: PropTypes.node.isRequired,
  positionReference: PropTypes.shape({}).isRequired,
  visible: PropTypes.bool.isRequired,
}

export default Portal
