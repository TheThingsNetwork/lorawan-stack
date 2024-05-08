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
import { Link } from 'react-router-dom'

import {
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
  IconX,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import SidebarContext from '@console/containers/sidebar/context'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './side-header.styl'

const SideHeader = ({ Logo }) => {
  const { onMinimizeToggle, isMinimized, onDrawerCloseClick } = useContext(SidebarContext)

  return (
    <div
      className={classnames(style.headerContainer, {
        [style.isMinimized]: isMinimized,
      })}
    >
      <Link to="/">
        <Logo className={[style.logo, { [style.minimizedLogo]: isMinimized }]} />
      </Link>
      <Button
        className={classnames(style.minimizeButton, 's:d-none')}
        icon={isMinimized ? IconLayoutSidebarLeftExpand : IconLayoutSidebarLeftCollapse}
        onClick={onMinimizeToggle}
        naked
      />
      <Button
        className={classnames(style.minimizeButton, 'd-none', 's:d-flex')}
        icon={IconX}
        onClick={onDrawerCloseClick}
        naked
      />
    </div>
  )
}

SideHeader.propTypes = {
  Logo: PropTypes.elementType.isRequired,
}

export default SideHeader
