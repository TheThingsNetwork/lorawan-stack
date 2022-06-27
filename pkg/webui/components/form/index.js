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

import React, { useCallback, useEffect, useState } from 'react'
import {
  yupToFormErrors,
  useFormikContext,
  validateYupSchema,
  useFormik,
  FormikProvider,
} from 'formik'
import scrollIntoView from 'scroll-into-view-if-needed'
import { defineMessages } from 'react-intl'
import { isPlainObject, isFunction, pick } from 'lodash'

import Notification from '@ttn-lw/components/notification'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PropTypes from '@ttn-lw/lib/prop-types'
import { ingestError } from '@ttn-lw/lib/errors/utils'

import FormField from './field'
import FormInfoField from './field/info'
import FormSubmit from './submit'
import FormCollapseSection from './section'
import FormSubTitle from './sub-title'
import FormFieldContainer from './field/container'

const m = defineMessages({
  submitFailed: 'Submit failed',
})

const Form = props => {
  const {
    children,
    className,
    disabled,
    enableReinitialize,
    error,
    errorTitle,
    formikRef,
    id,
    info,
    infoTitle,
    initialValues,
    onReset,
    onSubmit,
    validateOnBlur,
    validateOnChange,
    validateOnMount,
    validateSync,
    validationContext,
    validationSchema,
    stripUnusedFields,
  } = props

  const notificationRef = React.useRef()
  const [fieldRegistry, setFieldRegistry] = useState([])

  // Recreate the validation hook to allow passing down validation contexts.
  const validate = useCallback(
    values => {
      if (!validationSchema) {
        return {}
      }

      if (validateSync) {
        try {
          validateYupSchema(values, validationSchema, validateSync, validationContext)

          return {}
        } catch (err) {
          if (err.name === 'ValidationError') {
            return yupToFormErrors(err)
          }

          throw error
        }
      }

      return new Promise((resolve, reject) => {
        validateYupSchema(values, validationSchema, validateSync, validationContext).then(
          () => {
            resolve({})
          },
          err => {
            // Resolve yup errors, see https://jaredpalmer.com/formik/docs/migrating-v2#validate.
            if (err.name === 'ValidationError') {
              resolve(yupToFormErrors(err))
            } else {
              // Throw any other errors as it is not related to the validation process.
              reject(err)
            }
          },
        )
      })
    },
    [validationSchema, validateSync, validationContext, error],
  )

  // Initialize formik and get the formik context to provide to form children.
  const formik = useFormik({
    initialValues,
    validate,
    onSubmit,
    onReset,
    validateOnMount,
    validateOnBlur,
    validateSync,
    validateOnChange,
    enableReinitialize,
  })

  const {
    isSubmitting,
    isValid,
    handleSubmit: handleFormikSubmit,
    handleReset: handleFormikReset,
    registerField: registerFormikField,
    unregisterField: unregisterFormikField,
    setValues,
  } = formik

  // Recreate field registration, so the component can keep track of registered fields,
  // allowing automatic removal of unused field values from the value set if wished.
  const registerField = useCallback(
    (name, validate) => {
      registerFormikField(name, validate)
      if (stripUnusedFields) {
        setFieldRegistry(fieldRegistry => [...fieldRegistry, name])
      }
    },
    [registerFormikField, stripUnusedFields],
  )

  // Recreate field registration, so the component can keep track of registered fields,
  // allowing automatic removal of unused field values from the value set if wished.
  const unregisterField = useCallback(
    name => {
      if (stripUnusedFields) {
        unregisterFormikField(name)
      }
      setFieldRegistry(fieldRegistry => {
        const newRegistry = [...fieldRegistry]
        newRegistry.splice(newRegistry.indexOf(name), 1)

        return newRegistry
      })
    },
    [stripUnusedFields, unregisterFormikField],
  )

  // Recreate form submit handler to enable stripping values as well as error logging.
  const handleFormSubmit = useCallback(
    e => {
      try {
        if (stripUnusedFields) {
          // Clean up fields that have been unmounted at the time of submitting, if wished.
          setValues(values => pick(values, fieldRegistry))
        }
        handleFormikSubmit(e)
      } catch (error) {
        // Make sure all unhandled exceptions during submit are ingested.
        ingestError(error, { ingestedBy: 'FormSubmit' })

        throw error
      }
    },
    [fieldRegistry, handleFormikSubmit, setValues, stripUnusedFields],
  )

  // Connect the ref with the formik context to ensure compatibility with older form components.
  // NOTE: New components should not use the ref, but use the form context directly.
  // TODO: Remove this once all forms have been refactored to use context.
  if (isPlainObject(formikRef) && 'current' in formikRef) {
    formikRef.current = formik
  }

  useEffect(() => {
    // Scroll form notification into view if needed.
    if (error) {
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
  }, [error, isSubmitting, isValid])

  return (
    <FormikProvider
      value={{
        disabled,
        ...formik,
        registerField,
        unregisterField,
      }}
    >
      <form className={className} id={id} onSubmit={handleFormSubmit} onReset={handleFormikReset}>
        {(error || info) && (
          <div style={{ outline: 'none' }} ref={notificationRef} tabIndex="-1">
            {error && <ErrorNotification content={error} title={errorTitle} small />}
            {info && <Notification content={info} title={infoTitle} info small />}
          </div>
        )}
        {isFunction(children) ? children(formik) : children}
      </form>
    </FormikProvider>
  )
}

Form.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  disabled: PropTypes.bool,
  enableReinitialize: PropTypes.bool,
  error: PropTypes.error,
  errorTitle: PropTypes.message,
  formikRef: PropTypes.shape({ current: PropTypes.shape({}) }),
  id: PropTypes.string,
  info: PropTypes.message,
  infoTitle: PropTypes.message,
  initialValues: PropTypes.shape({}),
  onReset: PropTypes.func,
  onSubmit: PropTypes.func,
  stripUnusedFields: PropTypes.bool,
  validateOnBlur: PropTypes.bool,
  validateOnChange: PropTypes.bool,
  validateOnMount: PropTypes.bool,
  validateSync: PropTypes.bool,
  validationContext: PropTypes.shape({}),
  validationSchema: PropTypes.oneOfType([PropTypes.shape({}), PropTypes.func]),
}

Form.defaultProps = {
  className: undefined,
  disabled: false,
  enableReinitialize: false,
  error: undefined,
  errorTitle: m.submitFailed,
  info: undefined,
  infoTitle: undefined,
  formikRef: undefined,
  id: undefined,
  initialValues: undefined,
  onReset: () => null,
  onSubmit: () => null,
  stripUnusedFields: false,
  validateOnBlur: true,
  validateOnChange: false,
  validateOnMount: false,
  validateSync: true,
  validationContext: {},
  validationSchema: undefined,
}

Form.Field = FormField
Form.InfoField = FormInfoField
Form.Submit = FormSubmit
Form.CollapseSection = FormCollapseSection
Form.SubTitle = FormSubTitle
Form.FieldContainer = FormFieldContainer

export { Form as default, useFormikContext as useFormContext }
