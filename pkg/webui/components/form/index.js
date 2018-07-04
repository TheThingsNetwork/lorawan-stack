// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import { Formik } from 'formik'

import Field from '../field'
import Button from '../button'
import Notification from '../notification'

const InnerForm = function ({
  setFieldValue,
  setFieldTouched,
  handleSubmit,
  isSubmitting,
  errors,
  error,
  values,
  touched,
  children,
}) {

  const decoratedChildren = React.Children.map(children,
    function (Child) {
      if (Child.type === Field) {
        return React.cloneElement(Child, {
          setFieldValue,
          setFieldTouched,
          errors,
          values,
          touched,
          ...Child.props,
        })
      } else if (Child.type === Button && Child.props.type === 'submit') {
        return React.cloneElement(Child, {
          disabled: isSubmitting,
          ...Child.props,
        })
      }

      return Child
    })

  return (
    <form onSubmit={handleSubmit}>
      {error && (<Notification small error={error} />)}
      {decoratedChildren}
    </form>
  )
}

const Form = ({ children, error, ...rest }) => (
  <Formik {...rest} render={function (props) {
    return <InnerForm error={error} {...props}>{children}</InnerForm>
  }
  }
  />
)

export default Form
