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

import React, { useCallback, useState } from 'react'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './sidebar.styl'

const SidebarContext = React.createContext()

const SidebarContextProvider = ({ children }) => {
  // For the mobile side menu drawer functionality.
  const [isDrawerOpen, setIsDrawerOpen] = useState(false)
  const [isMinimized, setIsMinimized] = useState(false)
  const [isHovered, setIsHovered] = useState(false)

  const openDrawer = useCallback(() => {
    setIsDrawerOpen(true)
    document.body.classList.add(style.scrollLock)
  }, [setIsDrawerOpen])

  const closeDrawer = useCallback(() => {
    setIsDrawerOpen(false)
    document.body.classList.remove(style.scrollLock)
  }, [setIsDrawerOpen])

  const onMinimizeToggle = useCallback(async () => {
    setIsMinimized(prev => !prev)
    setIsDrawerOpen(false)
  }, [setIsMinimized])

  return (
    <SidebarContext.Provider
      value={{
        onMinimizeToggle,
        isMinimized,
        setIsMinimized,
        isDrawerOpen,
        setIsDrawerOpen,
        openDrawer,
        closeDrawer,
        isHovered,
        setIsHovered,
      }}
    >
      {children}
    </SidebarContext.Provider>
  )
}

SidebarContextProvider.propTypes = {
  children: PropTypes.node.isRequired,
}

export { SidebarContextProvider }

export default SidebarContext
