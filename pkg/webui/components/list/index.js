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

import PropTypes from '../../lib/prop-types'
import Message from '../../lib/components/message'
import sharedMessages from '../../lib/shared-messages'
import ListItem from './item'

import style from './list.styl'

@bind
class List extends React.PureComponent {
  renderItem (item, index) {
    const { rowKey, renderItem, size } = this.props

    let actualRowKey = null
    const rowKeyType = typeof rowKey
    if (rowKeyType === 'function') {
      actualRowKey = rowKey(item, index)
    } else if (rowKeyType === 'string') {
      actualRowKey = item[rowKey]
    } else if (item.key) {
      actualRowKey = item.key
    } else {
      actualRowKey = `list-item-${index}`
    }

    const renderedItem = renderItem(item, index)
    return React.cloneElement(renderedItem, {
      ...renderedItem.props,
      key: actualRowKey,
      className: classnames(
        renderedItem.props.className,
        style[`item-${size}`]
      ),
    })
  }

  renderItems () {
    const {
      items,
      emptyMessage,
      emptyMessageValues,
      children,
    } = this.props

    if (children) {
      return children
    }

    if (!items.length) {
      return (
        <Message
          className={style.listEmptyMessage}
          content={emptyMessage}
          values={emptyMessageValues}
        />
      )
    }

    return items.map((item, idx) => this.renderItem(item, idx))
  }

  render () {
    const {
      className,
      component: Component,
      bordered,
      items,
    } = this.props

    const cls = classnames(className, style.list, {
      [style.listBordered]: bordered,
      [style.listEmpty]: !items.length,
    })

    return (
      <Component className={cls}>
        {this.renderItems()}
      </Component>
    )
  }
}

List.propTypes = {
  size: PropTypes.oneOf([ 'small', 'default', 'large' ]),
  renderItem: PropTypes.func.isRequired,
  items: PropTypes.array,
  rowKey: PropTypes.oneOfType([ PropTypes.func, PropTypes.string ]),
  component: PropTypes.string,
  bordered: PropTypes.bool,
  emptyMessage: PropTypes.message,
  emptyMessageValues: PropTypes.object,
}

List.defaultProps = {
  component: 'ol',
  size: 'default',
  items: [],
  bordered: false,
  emptyMessage: sharedMessages.noMatch,
  emptyMessageValues: {},
}

List.Item = ListItem

export default List
