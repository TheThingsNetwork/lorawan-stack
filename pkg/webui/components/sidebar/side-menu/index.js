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

import React, { useRef } from 'react'
import classnames from 'classnames'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from './list'
import SideNavigationItem from './item'

import style from './side.styl'

const SideNavigation = ({ className, children }) => {
  const node = useRef()

  const navigationClassNames = classnames(className, style.navigation)

  return (
    <nav className={navigationClassNames} ref={node} data-test-id="navigation-sidebar">
      <SideNavigationList className={style.navigationList}>{children}</SideNavigationList>
    </nav>
  )
}

SideNavigation.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
}

SideNavigation.defaultProps = {
  className: undefined,
}

SideNavigation.Item = SideNavigationItem

export default SideNavigation
