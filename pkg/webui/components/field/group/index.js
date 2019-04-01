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

import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'
import getByPath from '../../../lib/get-by-path'
import Field, { FieldError } from '..'

import style from './group.styl'

class FieldGroup extends React.Component {
  render () {
    const {
      className,
      children,
      name,
      title,
      titleComponent = 'h4',
      error,
      value,
      disabled,
      setFieldValue,
      setFieldTouched,
      horizontal,
    } = this.props

    const fields = React.Children.map(children, function (Child) {
      if (React.isValidElement(Child) && Child.type === Field) {
        const fieldName = `${name}.${Child.props.name}`
        const fieldValue = getByPath(value, Child.props.name)
        return React.cloneElement(Child, {
          ...Child.props,
          name: fieldName,
          value: fieldValue,
          disabled,
          setFieldValue,
          setFieldTouched,
          horizontal,
        })
      }

      return Child
    })

    return (
      <div className={className}>
        <div className={style.header}>
          <Message
            className={style.headerTitle}
            component={titleComponent}
            content={title}
          />
          {error && <FieldError name={name} error={error} />}
        </div>
        {fields}
      </div>
    )
  }
}

FieldGroup.propTypes = {
  name: PropTypes.string.isRequired,
  title: PropTypes.message,
  errors: PropTypes.object,
}

export default FieldGroup
