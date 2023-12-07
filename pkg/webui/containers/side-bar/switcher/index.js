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

import React, { useCallback, useContext } from 'react'

import Switcher from '@ttn-lw/components/navigation/side-v2/switcher'

import SideBarContext from '@ttn-lw/containers/side-bar/context'

const SwitcherContainer = () => {
  const { layer, setLayer, isMinimized } = useContext(SideBarContext)

  const handleClick = useCallback(
    evt => {
      setLayer(evt.target.getAttribute('href'))
    },
    [setLayer],
  )

  return <Switcher layer={layer} isMinimized={isMinimized} onClick={handleClick} />
}

export default SwitcherContainer
