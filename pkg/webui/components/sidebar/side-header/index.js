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

import { IconLayoutSidebarLeftCollapse, IconX } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'

import SidebarContext from '@console/containers/sidebar/context'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './side-header.styl'

const SideHeader = ({ Logo }) => {
  const { onMinimizeToggle, isMinimized, closeDrawer } = useContext(SidebarContext)

  return (
    <div className={classnames(style.headerContainer)}>
      <Link to="/">
        <Logo className={classnames(style.logo)} />
      </Link>
      <Button
        className={classnames(style.minimizeButton, 'md-lg:d-none')}
        icon={isMinimized ? IconX : IconLayoutSidebarLeftCollapse}
        onClick={isMinimized ? closeDrawer : onMinimizeToggle}
        naked
      />
      <Button
        className={classnames(style.minimizeButton, 'd-none', 'md-lg:d-flex')}
        icon={IconX}
        onClick={closeDrawer}
        naked
      />
    </div>
  )
}

SideHeader.propTypes = {
  Logo: PropTypes.elementType.isRequired,
}

export default SideHeader
