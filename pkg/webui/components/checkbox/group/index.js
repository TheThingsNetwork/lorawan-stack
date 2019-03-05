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
import { defineMessages } from 'react-intl'

import sharedMessages from '../../../lib/shared-messages'
import Field from '../../field'
import Message from '../../../lib/components/message'
import Input from '../../input'
import PropTypes from '../../../lib/prop-types'
import debounce from '../../../lib/debounce'
import FieldGroup, { formatFieldName } from '../../field/group'

import style from './group.styl'

const m = defineMessages({
  checkAll: 'Check All',
  selections: 'Your Selections',
})

@bind
class CheckboxGroup extends React.Component {

  constructor (props) {
    super(props)

    this.state = {
      selected: [],
      searchQuery: '',
      allSelected: false,
    }

    if (props.searchDebounce) {
      const { debouncedFunction, cancel } = debounce(
        this.onSearchChange,
        props.searchDebounce
      )

      this.debouncedSearch = debouncedFunction
      this.debounceCancel = cancel
    } else {
      this.debouncedSearch = this.onSearchChange
      this.debounceCancel = () => null
    }
  }

  componentWillUnmount () {
    this.debounceCancel()
  }

  onSelectedChange (name, value) {
    const { onChange, children } = this.props
    const totalCount = React.Children.count(children)

    if (value) {
      this.setState(function (prev) {
        const selectedCount = prev.selected.length + 1

        return {
          selected: [ name, ...prev.selected ],
          allSelected: totalCount === selectedCount,
        }
      })
    } else {
      this.setState(prev => ({
        selected: prev.selected.filter(n => n !== name),
        allSelected: false,
      }))
    }

    if (onChange) {
      onChange(name, value)
    }
  }

  onTransferSelectedChange (name) {
    const { onChange } = this.props

    this.setState(prev => ({
      selected: prev.selected.filter(n => n !== name),
      allSelected: false,
    }), function () {
      if (onChange) {
        onChange(name, false)
      }
    })
  }

  onSearchChange (value) {
    this.setState({ searchQuery: value })
  }

  onSelectAll (_, value) {
    const { setValues, children, name: groupName } = this.props
    const { selected } = this.state

    const fields = React.Children.toArray(children)
    const newValues = {}
    const newSelected = value ? [ ...selected ] : []

    for (let i = 0; i < fields.length; i++) {
      const name = fields[i].props.name
      const fullName = formatFieldName(groupName, name)
      newValues[name] = value
      if (value && !selected.includes(fullName)) {
        newSelected.push(fullName)
      }
    }

    this.setState({
      selected: newSelected,
      allSelected: value,
    }, () => setValues({ [groupName]: newValues }))
  }

  get header () {
    const { transfer, selectAllTitle, selectAllName, children } = this.props
    const { selected, allSelected } = this.state

    const cls = classnames(style.header, {
      [style.headerWithTransfer]: transfer,
    })

    const selectedCount = selected.length
    const totalCount = React.Children.count(children)
    const indeterminateSelected = selectedCount !== 0 && !allSelected

    return (
      <div className={cls}>
        <div>
          <Field
            horizontal
            className={style.headerCheckbox}
            indeterminate={indeterminateSelected}
            type="checkbox"
            form={false}
            title={selectAllTitle}
            name={selectAllName}
            onChange={this.onSelectAll}
            value={allSelected}
          />
        </div>
        <span className={style.headerCountLabel}>
          {`(${selectedCount}/${totalCount})`}
        </span>
        { transfer && <Message content={m.selections} />}
      </div>
    )
  }

  get all () {
    const {
      search,
      form,
      onSearch,
      children,
      name,
    } = this.props
    const { selected } = this.state

    const onSelectedChange = this.onSelectedChange
    const { searchQuery } = this.state

    const searchInput = search
      ? (
        <div className={style.bodySearchWrapper}>
          <Input
            className={style.bodySearchInput}
            onChange={this.debouncedSearch}
            type="text"
            icon="search"
            name="search"
            placeholder="Search"
          />
        </div>
      ) : null

    return (
      <div className={style.bodyContentAll}>
        {searchInput}
        <List
          className={style.bodyContent}
        >
          {React.Children.toArray(children)
            .filter(field => onSearch(field.props, searchQuery))
            .map(function (field) {
              const fieldName = formatFieldName(name, field.props.name)
              const valueProp = form ? {} : { value: selected.includes(fieldName) }
              return React.cloneElement(field, {
                ...field.props,
                ...valueProp,
                onChange: onSelectedChange,
                type: 'checkbox',
                form,
              })
            })}
        </List>
      </div>
    )
  }

  get transfer () {
    const {
      transfer,
      form,
      children,
      name,
    } = this.props

    if (!transfer) {
      return null
    }

    const { selected } = this.state
    const onTransferSelectedChange = this.onTransferSelectedChange

    const fields = React.Children.toArray(children)
    const transferFields = selected.map(function (selectedName) {
      const field = fields.find(f => formatFieldName(name, f.props.name) === selectedName)
      return React.cloneElement(field, {
        ...field.props,
        key: field.props.name,
        type: 'checkbox',
        onChange: onTransferSelectedChange,
        value: true,
        form,
      })
    })

    return (
      <div className={style.bodyContentSelected}>
        <List
          className={style.bodyContent}
        >
          {transferFields}
        </List>
      </div>
    )
  }

  render () {
    const {
      className,
      name,
      title,
      errors,
      search,
      transfer,
    } = this.props

    const bodyClassNames = classnames(style.body, {
      [style.bodySearch]: search,
      [style.bodyWithTransfer]: transfer,
    })

    return (
      <FieldGroup
        name={name}
        title={title}
        errors={errors}
      >
        <div className={classnames(className, style.checkboxGroup)}>
          {this.header}
          <div className={bodyClassNames}>
            {this.all}
            {this.transfer}
          </div>
        </div>
      </FieldGroup>
    )
  }
}

const List = function (props) {
  const {
    className,
    children,
  } = props

  const fields = React.Children.map(children, function (Child) {
    return (
      <li
        className={style.bodyContentItem}
        key={name}
      >
        {React.cloneElement(Child, { ...Child.props })}
      </li>
    )
  })

  return (
    <ul className={className}>
      {fields}
    </ul>
  )
}

CheckboxGroup.propTypes = {
  form: PropTypes.bool,
  search: PropTypes.bool,
  transfer: PropTypes.bool,
  onSearch: PropTypes.func,
  searchDebounce: PropTypes.oneOfType([ PropTypes.number, PropTypes.bool ]),
  searchPlaceholder: PropTypes.message,
  name: PropTypes.string.isRequired,
  selectAllTitle: PropTypes.message,
  selectAllName: PropTypes.string,
}

CheckboxGroup.defaultProps = {
  form: true,
  search: false,
  transfer: false,
  onSearch: () => true,
  searchDebounce: 200,
  searchPlaceholder: sharedMessages.search,
  selectAllTitle: m.checkAll,
  selectAllName: 'check-all',
}

export default CheckboxGroup
