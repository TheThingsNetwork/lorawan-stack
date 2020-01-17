// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'
import ReactDom from 'react-dom'
import classnames from 'classnames'
import { Container } from 'react-grid-system'

import PropTypes from '../../lib/prop-types'

import style from './breadcrumbs.styl'

const Breadcrumbs = ({ className, breadcrumbs }) => (
  <nav className={classnames(className, style.breadcrumbs)}>
    {breadcrumbs.map(function(component, index) {
      return React.cloneElement(component, {
        key: index,
        isLast: index === breadcrumbs.length - 1,
      })
    })}
  </nav>
)

Breadcrumbs.propTypes = {
  /** A list of breadcrumb elements */
  breadcrumbs: PropTypes.arrayOf(PropTypes.oneOfType([PropTypes.func, PropTypes.element])),
  className: PropTypes.string,
}

Breadcrumbs.defaultProps = {
  breadcrumbs: undefined,
  className: undefined,
}

const PortalledBreadcrumbs = ({ className, ...rest }) => {
  // Breadcrumbs can be rendered into multiple containers
  const containers = document.querySelectorAll('.breadcrumbs, #breadcrumbs')
  if (containers.length) {
    const nodes = []
    containers.forEach(element => {
      nodes.push(
        ReactDom.createPortal(
          <div className={classnames(className, style.breadcrumbsContainer)}>
            <Container>
              <Breadcrumbs {...rest} />
            </Container>
          </div>,
          element,
        ),
      )
    })
    return nodes
  }

  return null
}

export { PortalledBreadcrumbs as default, Breadcrumbs }
