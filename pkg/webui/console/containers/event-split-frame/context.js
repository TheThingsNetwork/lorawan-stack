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

import React, { createContext, useState } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

// Define the shape of your context
const EventSplitFrameContext = createContext(null)

// Create a provider component
export const EventSplitFrameContextProvider = ({ children }) => {
  const [isOpen, setIsOpen] = useState(false)
  const [isActive, setIsActive] = useState(true)
  const [isMounted, setIsMounted] = useState(false)
  const [height, setHeight] = useState(20 * 14)

  return (
    <EventSplitFrameContext.Provider
      value={{
        isOpen,
        height: isOpen && isActive && isMounted ? height : 0,
        setIsOpen,
        setHeight,
        setIsActive,
        isActive,
        setIsMounted,
        isMounted,
      }}
    >
      {children}
    </EventSplitFrameContext.Provider>
  )
}

EventSplitFrameContextProvider.propTypes = {
  children: PropTypes.node.isRequired,
}

export default EventSplitFrameContext
