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
import bind from 'autobind-decorator'
import PropTypes from '../../../../lib/prop-types'

import style from './list.styl'

@bind
class SideNavigationList extends React.Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
    className: PropTypes.string,
    /** The depth of the current list starting at 0 for the root list */
    depth: PropTypes.number,
    /**
     * A flag specifying whether the side navigation list is expanded or not.
     * Applicable to nested lists.
     */
    isExpanded: PropTypes.bool,
  }

  static defaultProps = {
    className: undefined,
    depth: 0,
    isExpanded: false,
  }

  render() {
    const { children, className, isExpanded, depth } = this.props

    const isRoot = depth === 0
    const listClassNames = classnames(className, style.list, {
      [style.listNested]: !isRoot,
      [style.listExpanded]: isExpanded,
    })
    return <ul className={listClassNames}>{children}</ul>
  }
}

export default SideNavigationList
