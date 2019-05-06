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

import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import getByPath from '../../../lib/get-by-path'
import Field, { Field as PureField, FieldError } from '..'

import style from './group.styl'

class FieldGroup extends React.Component {
  render () {
    const {
      className: groupClassName,
      children,
      name: groupName,
      title,
      titleComponent = 'span',
      error,
      value,
      disabled,
      setFieldValue,
      setFieldTouched,
      horizontal,
      touched,
      columns,
      form = true,
    } = this.props
    const fields = React.Children.map(children, function (Child) {
      if (React.isValidElement(Child) && Child.type === Field || Child.type === PureField) {
        const { type, value: fieldValue, name: fieldName, className: fieldClassName } = Child.props
        const appliedProps = {}
        if (type === 'checkbox') {
          appliedProps.id = `${groupName}.${fieldName}`
          appliedProps.name = appliedProps.id
          if (form && value) {
            appliedProps.value = getByPath(value, fieldName)
          }
        } else if (type === 'radio') {
          appliedProps.name = groupName
          appliedProps.id = `${groupName}.${fieldValue}`
          if (form && value) {
            appliedProps.checked = value === fieldValue
          }
        }
        const classNames = classnames(style.field, fieldClassName, {
          [style.columns]: columns,
        })

        return React.cloneElement(Child, {
          ...Child.props,
          ...appliedProps,
          className: classNames,
          touches: groupName,
          disabled,
          setFieldValue,
          setFieldTouched,
          validateOnChange: true,
          horizontal,
        })
      }

      return Child
    })

    const classNames = classnames(style.container, groupClassName, {
      [style.horizontal]: horizontal,
      [style.disabled]: disabled,
    })

    return (
      <div className={classNames}>
        <Message
          className={style.headerTitle}
          component={titleComponent}
          content={title}
        />
        <div
          className={style.fields}
        >
          {fields}
          {error && touched && <FieldError className={style.error} name={name} error={error} />}
        </div>
      </div>
    )
  }
}

FieldGroup.propTypes = {
  name: PropTypes.string.isRequired,
  title: PropTypes.message,
  error: PropTypes.error,
  horizontal: PropTypes.bool,
  columns: PropTypes.bool,
}

export default FieldGroup
