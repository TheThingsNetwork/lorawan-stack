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

import React, { useContext } from 'react'
import classnames from 'classnames'

import LAYOUT from '@ttn-lw/constants/layout'

import Button from '@ttn-lw/components/button-v2'

import SidebarContext from '@console/containers/side-bar/context'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './side-header.styl'

const getViewportWidth = () =>
  Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)

const SideHeader = ({ logo }) => {
  const { onMinimizeToggle, isMinimized } = useContext(SidebarContext)

  const viewportWidth = getViewportWidth()
  const isMobile = viewportWidth <= LAYOUT.BREAKPOINTS.M

  return (
    <div
      className={classnames('d-flex', 'j-between', {
        'direction-column': isMinimized,
        'gap-cs-xs': isMinimized,
        'al-center': isMinimized,
      })}
    >
      <img
        {...logo}
        className={classnames({ 'w-50': !isMinimized, [style.miniLogo]: isMinimized })}
      />
      {!isMobile && <Button icon="left_panel_close" naked onClick={onMinimizeToggle} />}
    </div>
  )
}

SideHeader.propTypes = {
  logo: PropTypes.shape({
    src: PropTypes.string.isRequired,
    alt: PropTypes.string.isRequired,
  }).isRequired,
}

export default SideHeader
