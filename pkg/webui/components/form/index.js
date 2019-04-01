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
import { Formik } from 'formik'
import bind from 'autobind-decorator'

import Field from '../field'
import FieldGroup from '../field/group'
import Button from '../button'
import Notification from '../notification'
import PropTypes from '../../lib/prop-types'
import getByPath from '../../lib/get-by-path'

@bind
class InnerForm extends React.Component {

  componentDidUpdate (prev) {
    const {
      loading,
      setSubmitting,
      setStatus,
      setTouched,
      status = {},
      values,
      initialValues,
      error,
      mapErrorsToFields,
    } = this.props

    if (prev.loading && !loading) {
      setSubmitting(loading)
    }

    // add field errors from the outside
    if (prev.error !== error) {
      const apiFieldErrors = fieldErrors(mapErrorsToFields, error)
      const { errors, ...restStatus } = status

      if (apiFieldErrors) {
        const forceTouched = Object.keys(apiFieldErrors)
          .reduce((acc, curr) => ({ ...acc, [curr]: true }), {})
        setTouched(forceTouched)
        setStatus({ errors: apiFieldErrors, ...restStatus })
      } else {
        setStatus({ formError: error })
      }
    }

    // remove errors from the outside on value change
    if (status.errors && prev.values !== values) {
      const { errors, ...restStatus } = status
      const errs = { ...errors }
      const forceTouched = {}

      for (const fieldName in errs) {
        const err = status.errors[fieldName]
        if (err && values[fieldName] !== initialValues[fieldName]) {
          delete errs[fieldName]
          forceTouched[fieldName] = true
        }
      }

      setTouched(forceTouched)
      setStatus({ errors: errs, ...restStatus })
    }
  }

  render () {
    const {
      setFieldValue,
      setFieldTouched,
      handleSubmit,
      handleReset,
      isSubmitting,
      isValid,
      errors,
      error,
      info,
      values,
      touched,
      children,
      horizontal,
      submitEnabledWhenInvalid,
      resetEnabledAlways,
      validateOnBlur,
      validateOnChange,
      dirty,
      status = {},
    } = this.props

    const formError = status.formError || false
    const serverErrors = status.errors || {}
    const clientErrors = errors
    const combinedErrors = { ...serverErrors, ...clientErrors }

    const decoratedChildren = recursiveMap(children,
      function (Child) {
        if (Child.type === Field) {
          const { name, value: originalValue } = Child.props
          const value = name ? getByPath(values, name) : originalValue
          const fieldError = getByPath(combinedErrors, name)
          const fieldTouched = getByPath(touched, name)

          return React.cloneElement(Child, {
            setFieldValue,
            setFieldTouched,
            error: fieldError,
            touched: fieldTouched,
            value,
            horizontal,
            submitEnabledWhenInvalid,
            validateOnBlur,
            validateOnChange,
            ...Child.props,
          })
        } else if (Child.type === Button) {
          if (Child.props.type === 'submit') {
            return React.cloneElement(Child, {
              ...Child.props,
              disabled: isSubmitting || !submitEnabledWhenInvalid && !isValid,
              busy: isSubmitting,
            })
          } else if (Child.props.type === 'reset') {
            const disabled = isSubmitting || (resetEnabledAlways ? false : !dirty)
            return React.cloneElement(Child, {
              ...Child.props,
              disabled,
              onClick: handleReset,
            })
          }
        } else if (Child.type === FieldGroup) {
          const { name } = Child.props
          const value = name && getByPath(values, name)
          const groupError = getByPath(combinedErrors, name)
          const groupTouched = getByPath(touched, name)
          return React.cloneElement(Child, {
            ...Child.props,
            setFieldValue,
            setFieldTouched,
            touched: groupTouched,
            error: groupError,
            value,
            horizontal,
          })
        }

        return Child
      })

    return (
      <form onSubmit={handleSubmit}>
        {formError && (<Notification small error={error} />)}
        {info && (<Notification small info={info} />)}
        {decoratedChildren}
      </form>
    )
  }
}

const formRender = ({ children, ...rest }) => function (props) {
  return (
    <InnerForm
      {...props}
      {...rest}
    >
      {children}
    </InnerForm>
  )
}

class Form extends React.Component {
  render () {
    const {
      children,
      error,
      info,
      loading,
      horizontal,
      submitEnabledWhenInvalid,
      resetEnabledAlways,
      validateOnBlur = true,
      validateOnChange = false,
      mapErrorsToFields = {},
      formikRef,
      ...rest
    } = this.props

    return (
      <Formik
        {...rest}
        ref={formikRef}
        validateOnBlur={validateOnBlur}
        validateOnChange={validateOnChange}
        render={formRender({
          children,
          error,
          info,
          horizontal,
          submitEnabledWhenInvalid,
          resetEnabledAlways,
          loading,
          mapErrorsToFields,
        })}
      />
    )
  }
}

function recursiveMap (children, fn) {
  return React.Children.map(children, function (Child) {
    if (!React.isValidElement(Child)) {
      return Child
    }

    let child = Child
    if (child.props.children && child.type !== FieldGroup) {
      child = React.cloneElement(child, {
        children: recursiveMap(child.props.children, fn),
      })
    }

    return fn(child)
  })
}

const fieldErrors = function (definition, error) {
  // stack custom errors
  if (typeof error === 'object' && error.details) {
    const formatted = {}

    error.details.forEach(function (detail) {
      const fieldName = definition[detail.name]
      if (fieldName) {
        const err = {}
        err.id = error.message.split(' ')[0]
        err.defaultMessage = error.details[0].message_format || error.message.replace(/^.*\s/, '')
        err.values = error.details[0].attribute

        formatted[fieldName] = err
      }
    })

    return Object.keys(formatted).length ? formatted : null
  }
}

Form.propTypes = {
  /** An error message belonging to the form */
  error: PropTypes.error,
  /** Whether the form fields should be displayed in horizontal style */
  horizontal: PropTypes.bool,
  /** Whether the submit button stays enabled also when the form data is not
   * not yet valid */
  submitEnabledWhenInvalid: PropTypes.bool,
  /** The flag specifying whether the form is in the loading state */
  loading: PropTypes.bool,
  /** Field name to stack error name mappings, e.g. { id: 'invalid_id' } */
  mapErrorsToFields: PropTypes.object,
  /** Whether the reset/cancel buttons stays enabled also when the form is not dirty */
  resetEnabledAlways: PropTypes.bool,
  /** A reference property passed to the formik component */
  formikRef: PropTypes.shape({ current: PropTypes.instanceOf(Formik) }),
}

export default Form
