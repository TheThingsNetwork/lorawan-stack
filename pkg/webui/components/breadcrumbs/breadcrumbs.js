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
import classnames from 'classnames'
import { Container } from 'react-grid-system'
import ReactDom from 'react-dom'

import Link from '@ttn-lw/components/link'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './breadcrumbs.styl'

const Breadcrumbs = ({ className, breadcrumbs }) => (
  <div className={classnames(className, style.breadcrumbs, 'd-flex', 'al-center', 'gap-cs-xs')}>
    {breadcrumbs.map((b, index) =>
      index !== breadcrumbs.length - 1 ? (
        <React.Fragment key={b.id}>
          <Link to={b.path} secondary className={style.link}>
            <Message content={b.content} />
          </Link>
          <Icon icon="arrow_forward_ios" small className={style['arrow-icon']} />
        </React.Fragment>
      ) : (
        <Message
          key={b.id}
          content={b.content}
          component="p"
          className={classnames(style.last, 'm-0')}
        />
      ),
    )}
  </div>
)

Breadcrumbs.propTypes = {
  /** An array of breadcrumbs. */
  breadcrumbs: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      path: PropTypes.string.isRequired,
      content: PropTypes.string.isRequired,
    }),
  ).isRequired,
  className: PropTypes.string,
}

Breadcrumbs.defaultProps = {
  className: undefined,
}

const PortalledBreadcrumbs = ({ className, ...rest }) => {
  // Breadcrumbs can be rendered into multiple containers.
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

PortalledBreadcrumbs.propTypes = {
  className: PropTypes.string,
}

PortalledBreadcrumbs.defaultProps = {
  className: undefined,
}

export { PortalledBreadcrumbs as default, Breadcrumbs }
