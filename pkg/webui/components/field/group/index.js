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
import { FieldError } from '..'

import style from './group.styl'

class FieldGroup extends React.Component {
  render () {
    const {
      className,
      children,
      name,
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
    } = this.props

    const fields = React.Children.map(children, function (Child) {

      if (React.isValidElement(Child) && Child.type.name === 'Field') {
        const { type, value: fieldValue } = Child.props
        let id, fieldName
        const mapValues = {}
        if (type === 'checkbox') {
          fieldName = id
          id = `${name}.${fieldName}`
          mapValues.value = getByPath(value, Child.props.name)
        } else if (type === 'radio') {
          fieldName = name
          id = `${name}.${fieldValue}`
          mapValues.checked = value === fieldValue
        }
        const classNames = classnames(style.field, className, {
          [style.columns]: columns,
        })

        return React.cloneElement(Child, {
          ...Child.props,
          ...mapValues,
          className: classNames,
          name: fieldName,
          value: fieldValue,
          touches: name,
          disabled,
          setFieldValue,
          setFieldTouched,
          validateOnChange: true,
          horizontal,
          id,
        })
      }

      return Child
    })

    const classNames = classnames(style.container, className, {
      [style.horizontal]: horizontal,
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
  errors: PropTypes.object,
  horizontal: PropTypes.bool,
  columns: PropTypes.bool,
}

export default FieldGroup
