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
  renderItem(item, index) {
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
      className: classnames(renderedItem.props.className, {
        [style[`item-${size}`]]: size !== 'none',
      }),
    })
  }

  get header() {
    const { header, size } = this.props

    if (!header) {
      return null
    }

    return <div className={classnames(style.header, style[`item-${size}`])}>{header}</div>
  }

  get footer() {
    const { footer, size } = this.props

    if (!footer) {
      return null
    }

    return <div className={classnames(style.footer, style[`item-${size}`])}>{footer}</div>
  }

  renderItems() {
    const { items, emptyMessage, emptyMessageValues, children } = this.props

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

  render() {
    const { className, component: Component, bordered, items, listClassName } = this.props

    const cls = classnames(className, style.wrapper, {
      [style.listBordered]: bordered,
    })

    const listCls = classnames(style.list, listClassName, {
      [style.listEmpty]: !items.length,
    })

    return (
      <div className={cls}>
        {this.header}
        <Component className={listCls}>{this.renderItems()}</Component>
        {this.footer}
      </div>
    )
  }
}

List.propTypes = {
  size: PropTypes.oneOf(['small', 'default', 'large', 'none']),
  renderItem: PropTypes.func,
  items: PropTypes.array,
  rowKey: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  component: PropTypes.oneOf(['ol', 'ul']),
  bordered: PropTypes.bool,
  emptyMessage: PropTypes.message,
  emptyMessageValues: PropTypes.object,
  header: PropTypes.node,
  listClassName: PropTypes.string,
  footer: PropTypes.node,
}

List.defaultProps = {
  component: 'ol',
  size: 'default',
  items: [],
  bordered: false,
  emptyMessage: sharedMessages.noMatch,
  emptyMessageValues: {},
  header: null,
  renderItem: () => null,
  footer: null,
}

List.Item = ListItem

export default List
