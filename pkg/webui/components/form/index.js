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

/* eslint-disable react/sort-prop-types */
import React from 'react'
import { Formik, yupToFormErrors, useFormikContext, validateYupSchema } from 'formik'
import bind from 'autobind-decorator'
import scrollIntoView from 'scroll-into-view-if-needed'
import classnames from 'classnames'

import Notification from '@ttn-lw/components/notification'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PropTypes from '@ttn-lw/lib/prop-types'

import FormContext from './context'
import FormField from './field'
import FormInfoField from './field/info'
import FormSubmit from './submit'
import FormCollapseSection from './section'
import FormSubTitle from './sub-title'

import style from './form.styl'

class InnerForm extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
    className: PropTypes.string,
    formError: PropTypes.error,
    formErrorTitle: PropTypes.message,
    formInfo: PropTypes.message,
    formInfoTitle: PropTypes.message,
    handleSubmit: PropTypes.func.isRequired,
    isSubmitting: PropTypes.bool.isRequired,
    isValid: PropTypes.bool.isRequired,
  }

  static defaultProps = {
    className: undefined,
    formError: undefined,
    formErrorTitle: undefined,
    formInfo: undefined,
    formInfoTitle: undefined,
  }

  constructor(props) {
    super(props)
    this.notificationRef = React.createRef()
  }

  componentDidUpdate(prevProps) {
    const { formError, isSubmitting, isValid } = this.props
    const { isSubmitting: prevIsSubmitting, formError: prevFormError } = prevProps

    // Scroll form notification into view if needed.
    if (formError && !prevFormError) {
      scrollIntoView(this.notificationRef.current, { behavior: 'smooth' })
      this.notificationRef.current.focus({ preventScroll: true })
    }

    // Scroll invalid fields into view if needed and focus them.
    if (prevIsSubmitting && !isSubmitting && !isValid) {
      const firstErrorNode = document.querySelectorAll('[data-needs-focus="true"]')[0]
      if (firstErrorNode) {
        scrollIntoView(firstErrorNode, { behavior: 'smooth' })
        firstErrorNode.querySelector('input,textarea').focus({ preventScroll: true })
      }
    }
  }

  render() {
    const {
      className,
      children,
      formError,
      formErrorTitle,
      formInfo,
      formInfoTitle,
      handleSubmit,
      ...rest
    } = this.props

    return (
      <form className={classnames(style.container, className)} onSubmit={handleSubmit}>
        {(formError || formInfo) && (
          <div style={{ outline: 'none' }} ref={this.notificationRef} tabIndex="-1">
            {formError && <ErrorNotification content={formError} title={formErrorTitle} small />}
            {formInfo && <Notification content={formInfo} title={formInfoTitle} info small />}
          </div>
        )}
        <FormContext.Provider
          value={{
            ...rest,
          }}
        >
          {children}
        </FormContext.Provider>
      </form>
    )
  }
}

const formRenderer = ({ children, ...rest }) =>
  function (renderProps) {
    const { className, error, errorTitle, info, infoTitle, disabled } = rest
    const { handleSubmit, ...restFormikProps } = renderProps

    return (
      <InnerForm
        className={className}
        formError={error}
        formErrorTitle={errorTitle}
        formInfo={info}
        formInfoTitle={infoTitle}
        handleSubmit={handleSubmit}
        disabled={disabled}
        {...restFormikProps}
      >
        {children}
      </InnerForm>
    )
  }

class Form extends React.PureComponent {
  static propTypes = {
    enableReinitialize: PropTypes.bool,
    formikRef: PropTypes.shape({ current: PropTypes.any }),
    initialValues: PropTypes.shape({}),
    onReset: PropTypes.func,
    onSubmit: PropTypes.func.isRequired,
    validateOnMount: PropTypes.bool,
    validateOnBlur: PropTypes.bool,
    validateOnChange: PropTypes.bool,
    validationSchema: PropTypes.oneOfType([PropTypes.shape({}), PropTypes.func]),
    validationContext: PropTypes.shape({}),
    validateSync: PropTypes.bool,
  }

  static defaultProps = {
    enableReinitialize: false,
    formikRef: undefined,
    initialValues: undefined,
    onReset: () => null,
    validateOnBlur: true,
    validateOnMount: false,
    validateOnChange: false,
    validationSchema: undefined,
    validationContext: {},
    validateSync: true,
  }

  @bind
  validate(values) {
    const { validationSchema, validationContext, validateSync } = this.props

    if (!validationSchema) {
      return {}
    }

    if (validateSync) {
      try {
        validateYupSchema(values, validationSchema, validateSync, validationContext)

        return {}
      } catch (error) {
        if (error.name === 'ValidationError') {
          return yupToFormErrors(error)
        }

        throw error
      }
    }

    return new Promise((resolve, reject) => {
      validateYupSchema(values, validationSchema, validateSync, validationContext).then(
        () => {
          resolve({})
        },
        error => {
          // Resolve yup errors, see https://jaredpalmer.com/formik/docs/migrating-v2#validate.
          if (error.name === 'ValidationError') {
            resolve(yupToFormErrors(error))
          } else {
            // Throw any other errors as it is not related to the validation process.
            reject(error)
          }
        },
      )
    })
  }

  render() {
    const {
      onSubmit,
      onReset,
      initialValues,
      validateOnBlur,
      validateOnChange,
      validationSchema,
      validationContext,
      validateOnMount,
      formikRef,
      enableReinitialize,
      ...rest
    } = this.props

    return (
      <Formik
        innerRef={formikRef}
        validate={this.validate}
        onSubmit={onSubmit}
        onReset={onReset}
        validateOnMount={validateOnMount}
        initialValues={initialValues}
        validateOnBlur={validateOnBlur}
        validateOnChange={validateOnChange}
        enableReinitialize={enableReinitialize}
      >
        {formRenderer(rest)}
      </Formik>
    )
  }
}

Form.Field = FormField
Form.InfoField = FormInfoField
Form.Submit = FormSubmit
Form.CollapseSection = FormCollapseSection
Form.SubTitle = FormSubTitle

export { Form as default, useFormikContext as useFormContext }
