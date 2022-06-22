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
import React, { useCallback, useEffect } from 'react'
import { Formik, yupToFormErrors, useFormikContext, validateYupSchema } from 'formik'
import bind from 'autobind-decorator'
import scrollIntoView from 'scroll-into-view-if-needed'
import { defineMessages } from 'react-intl'

import Notification from '@ttn-lw/components/notification'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PropTypes from '@ttn-lw/lib/prop-types'
import { ingestError } from '@ttn-lw/lib/errors/utils'

import FormContext from './context'
import FormField from './field'
import FormInfoField from './field/info'
import FormSubmit from './submit'
import FormCollapseSection from './section'
import FormSubTitle from './sub-title'
import FormFieldContainer from './field/container'

const m = defineMessages({
  submitFailed: 'Submit failed',
})

const InnerForm = props => {
  const {
    formError,
    isSubmitting,
    isValid,
    className,
    children,
    formErrorTitle,
    formInfo,
    formInfoTitle,
    handleSubmit,
    id,
    ...rest
  } = props
  const notificationRef = React.useRef()

  useEffect(() => {
    // Scroll form notification into view if needed.
    if (formError) {
      scrollIntoView(notificationRef.current, { behavior: 'smooth' })
      notificationRef.current.focus({ preventScroll: true })
    }

    // Scroll invalid fields into view if needed and focus them.
    if (!isSubmitting && !isValid) {
      const firstErrorNode = document.querySelectorAll('[data-needs-focus="true"]')[0]
      if (firstErrorNode) {
        scrollIntoView(firstErrorNode, { behavior: 'smooth' })
        firstErrorNode.querySelector('input,textarea,canvas,video').focus({ preventScroll: true })
      }
    }
  }, [formError, isSubmitting, isValid])

  return (
    <form className={className} onSubmit={handleSubmit} id={id}>
      {(formError || formInfo) && (
        <div style={{ outline: 'none' }} ref={notificationRef} tabIndex="-1">
          {formError && <ErrorNotification content={formError} title={formErrorTitle} small />}
          {formInfo && <Notification content={formInfo} title={formInfoTitle} info small />}
        </div>
      )}
      <FormContext.Provider
        value={{
          formError,
          ...rest,
        }}
      >
        {children}
      </FormContext.Provider>
    </form>
  )
}
InnerForm.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  id: PropTypes.string,
  formError: PropTypes.error,
  formErrorTitle: PropTypes.message,
  formInfo: PropTypes.message,
  formInfoTitle: PropTypes.message,
  handleSubmit: PropTypes.func.isRequired,
  isSubmitting: PropTypes.bool.isRequired,
  isValid: PropTypes.bool.isRequired,
}

InnerForm.defaultProps = {
  className: undefined,
  id: undefined,
  formInfo: undefined,
  formInfoTitle: undefined,
  formError: undefined,
  formErrorTitle: m.submitFailed,
}

const formRenderer =
  ({ children, ...rest }) =>
  renderProps => {
    const { className, error, errorTitle, info, infoTitle, disabled, id } = rest
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
        id={id}
        {...restFormikProps}
      >
        {children}
      </InnerForm>
    )
  }

const Form = props => {
  const {
    onReset,
    initialValues,
    validateOnBlur,
    validateOnChange,
    validationSchema,
    validationContext,
    validateOnMount,
    formikRef,
    enableReinitialize,
    onSubmit,
    validateSync,
    ...rest
  } = props

  const handleSubmit = useCallback(
    async (...args) => {
      try {
        return await onSubmit(...args)
      } catch (error) {
        // Make sure all unhandled exceptions during submit are ingested.
        ingestError(error, { ingestedBy: 'FormSubmit' })

        throw error
      }
    },
    [onSubmit],
  )

  const validate = useEffect(
    values => {
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
    },
    [validationSchema, validateSync, validationContext],
  )

  return (
    <Formik
      innerRef={formikRef}
      validate={validate}
      onSubmit={handleSubmit}
      onReset={onReset}
      validateOnMount={validateOnMount}
      initialValues={initialValues}
      validateOnBlur={validateOnBlur}
      validateSync={validateSync}
      validateOnChange={validateOnChange}
      enableReinitialize={enableReinitialize}
    >
      {formRenderer(rest)}
    </Formik>
  )
}

Form.propTypes = {
  enableReinitialize: PropTypes.bool,
  formikRef: PropTypes.shape({ current: PropTypes.shape({}) }),
  initialValues: PropTypes.shape({}),
  onReset: PropTypes.func,
  onSubmit: PropTypes.func,
  validateOnMount: PropTypes.bool,
  validateOnBlur: PropTypes.bool,
  validateOnChange: PropTypes.bool,
  validationSchema: PropTypes.oneOfType([PropTypes.shape({}), PropTypes.func]),
  validationContext: PropTypes.shape({}),
  validateSync: PropTypes.bool,
}

Form.defaultProps = {
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
  onSubmit: () => null,
}

Form.Field = FormField
Form.InfoField = FormInfoField
Form.Submit = FormSubmit
Form.CollapseSection = FormCollapseSection
Form.SubTitle = FormSubTitle
Form.FieldContainer = FormFieldContainer

export { Form as default, useFormikContext as useFormContext }
