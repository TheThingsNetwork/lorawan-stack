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

import React, { useRef, useEffect, useContext } from 'react'
import classnames from 'classnames'

import SidebarContext from '@console/containers/sidebar/context'

import PropTypes from '@ttn-lw/lib/prop-types'

import SideNavigationList from './list'
import SideNavigationItem from './item'

import style from './side.styl'

const SideNavigation = ({ className, children }) => {
  const node = useRef()
  const { isMinimized } = useContext(SidebarContext)

  const navigationClassNames = classnames(className, style.navigation, {
    [style.isMinimized]: isMinimized,
  })

  // Add a scroll gradient to the navigation sidebar.
  useEffect(() => {
    const container = node?.current
    const handleScroll = () => {
      if (container) {
        const { scrollTop, scrollHeight, clientHeight } = container
        const scrollable = scrollHeight - clientHeight
        const scrollGradientTop = container.querySelector(`.${style.scrollGradientTop}`)
        const scrollGradientBottom = container.querySelector(`.${style.scrollGradientBottom}`)
        const fadeHeight = 20 // Height in pixels where the gradient starts to appear.

        if (scrollGradientTop) {
          const opacity = scrollTop < fadeHeight ? scrollTop / fadeHeight : 1
          scrollGradientTop.style.opacity = opacity
        }

        if (scrollGradientBottom) {
          const scrollEnd = scrollable - fadeHeight
          const opacity = scrollTop < scrollEnd ? 1 : (scrollable - scrollTop) / fadeHeight
          scrollGradientBottom.style.opacity = opacity
        }
      }
    }

    handleScroll()

    if (container) {
      container.addEventListener('scroll', handleScroll)
      window.addEventListener('resize', handleScroll)
    }

    return () => {
      if (container) {
        container.removeEventListener('scroll', handleScroll)
        window.removeEventListener('resize', handleScroll)
      }
    }
  }, [node])

  return (
    <nav className={navigationClassNames} ref={node} data-test-id="navigation-sidebar">
      <div className={style.scrollGradientTop} />
      <SideNavigationList className={style.navigationList}>{children}</SideNavigationList>
      <div className={style.scrollGradientBottom} />
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
