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
const Portal = ({ children, positionReferenceId, elementId, place }) => {
  const [portalContainer, setPortalContainer] = React.useState(undefined)

  useEffect(() => {
    const div = document.createElement('div')
    div.id = elementId
    document.body.appendChild(div)
    const element = document.getElementById(elementId)
    setPortalContainer(element)
    const triggerButton = document.getElementById(positionReferenceId)
    const rect = triggerButton.getBoundingClientRect()
    const elementLeft = rect.left + window.scrollX
    const elementTop = rect.top + window.scrollY
    element.style.position = 'absolute'
    element.style.left = place.includes('under')
      ? `${elementLeft}px`
      : `${elementLeft + rect.width}px`
    element.style.top = `${elementTop + rect.height}px`
    return () => {
      document.body.removeChild(document.getElementById(elementId))
      setPortalContainer(undefined)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return Boolean(portalContainer) && DOM.createPortal(children, document.getElementById(elementId))
}

Portal.propTypes = {
  children: PropTypes.node.isRequired,
  elementId: PropTypes.string.isRequired,
  positionReferenceId: PropTypes.string.isRequired,
}

export default Portal
