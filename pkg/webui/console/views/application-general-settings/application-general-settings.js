// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { isEqual } from 'lodash'

import DeleteModalButton from '@ttn-lw/components/delete-modal-button'
import PageTitle from '@ttn-lw/components/page-title'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import SubmitBar from '@ttn-lw/components/submit-bar'
import KeyValueMap from '@ttn-lw/components/key-value-map'
import Checkbox from '@ttn-lw/components/checkbox'

import Require from '@console/lib/components/require'

import Yup from '@ttn-lw/lib/yup'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { mayDeleteApplication } from '@console/lib/feature-checks'
import {
  attributeValidCheck,
  attributeTooShortCheck,
  attributeKeyTooLongCheck,
  attributeValueTooLongCheck,
} from '@console/lib/attributes'

import { mapFormValuesToApplication, mapApplicationToFormValues } from './mapping'

const m = defineMessages({
  basics: 'Basics',
  deleteApp: 'Delete application',
  modalWarning:
    'Are you sure you want to delete "{appName}"? This action cannot be undone and it will not be possible to reuse the application ID.',
  updateSuccess: 'Application updated',
})

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(3, Yup.passValues(sharedMessages.validateTooShort))
    .max(50, Yup.passValues(sharedMessages.validateTooLong)),
  description: Yup.string().max(150, Yup.passValues(sharedMessages.validateTooLong)),
  attributes: Yup.array()
    .max(10, Yup.passValues(sharedMessages.attributesValidateTooMany))
    .test(
      'has no empty string values',
      sharedMessages.attributesValidateRequired,
      attributeValidCheck,
    )
    .test(
      'has key length longer than 2',
      sharedMessages.attributeKeyValidateTooShort,
      attributeTooShortCheck,
    )
    .test(
      'has key length less than 36',
      sharedMessages.attributeKeyValidateTooLong,
      attributeKeyTooLongCheck,
    )
    .test(
      'has value length less than 200',
      sharedMessages.attributeValueValidateTooLong,
      attributeValueTooLongCheck,
    ),
  skip_payload_crypto: Yup.boolean(),
})

const ApplicationGeneralSettings = props => {
  const {
    appId,
    application,
    shouldConfirmDelete,
    mayPurge,
    link,
    mayViewLink,
    updateApplication,
    updateApplicationLink,
    deleteApplication,
    onDeleteSuccess,
  } = props

  const [error, setError] = useState()
  const initialValues = mapApplicationToFormValues({ ...application, ...link })

  const handleSubmit = useCallback(
    async (values, { resetForm, setSubmitting }) => {
      setError(undefined)

      const appValues = mapFormValuesToApplication(values)
      if (isEqual(application.attributes || {}, appValues.attributes)) {
        delete appValues.attributes
      }

      const changed = diff(application, appValues)

      // If there is a change in attributes, copy all attributes so they don't get
      // overwritten.
      const update =
        'attributes' in changed
          ? {
              ...changed,
              attributes: appValues.attributes,
            }
          : changed

      const {
        ids: { application_id },
      } = application

      try {
        const { skip_payload_crypto, ...applicationUpdate } = update
        const linkUpdate = { skip_payload_crypto }
        await updateApplication(application_id, applicationUpdate)
        await updateApplicationLink(application_id, linkUpdate)
        resetForm({ values })
        toast({
          title: application_id,
          message: m.updateSuccess,
          type: toast.types.SUCCESS,
        })
      } catch (error) {
        setSubmitting(false)
        setError(error)
      }
    },
    [application, updateApplication, updateApplicationLink],
  )
  const handleDelete = useCallback(
    async shouldPurge => {
      setError(undefined)

      try {
        await deleteApplication(appId, shouldPurge)
        onDeleteSuccess()
      } catch (error) {
        setError(error)
      }
    },
    [appId, deleteApplication, onDeleteSuccess],
  )

  return (
    <Container>
      <PageTitle title={sharedMessages.generalSettings} />
      <Row>
        <Col lg={8} md={12}>
          <Form
            error={error}
            onSubmit={handleSubmit}
            initialValues={initialValues}
            validationSchema={validationSchema}
          >
            <Form.Field
              title={sharedMessages.appId}
              name="ids.application_id"
              required
              component={Input}
              disabled
            />
            <Form.Field title={sharedMessages.name} name="name" component={Input} />
            <Form.Field
              title={sharedMessages.description}
              type="textarea"
              name="description"
              component={Input}
            />
            <Form.Field
              name="attributes"
              title={sharedMessages.attributes}
              keyPlaceholder={sharedMessages.key}
              valuePlaceholder={sharedMessages.value}
              addMessage={sharedMessages.addAttributes}
              component={KeyValueMap}
              description={sharedMessages.attributeDescription}
            />
            {mayViewLink && (
              <Form.Field
                autoFocus
                title={sharedMessages.skipCryptoTitle}
                name="skip_payload_crypto"
                description={sharedMessages.skipCryptoDescription}
                component={Checkbox}
              />
            )}
            <SubmitBar>
              <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
              <Require featureCheck={mayDeleteApplication}>
                <DeleteModalButton
                  message={m.deleteApp}
                  entityId={appId}
                  entityName={application.name}
                  onApprove={handleDelete}
                  shouldConfirm={shouldConfirmDelete}
                  mayPurge={mayPurge}
                />
              </Require>
            </SubmitBar>
          </Form>
        </Col>
      </Row>
    </Container>
  )
}

ApplicationGeneralSettings.propTypes = {
  appId: PropTypes.string.isRequired,
  application: PropTypes.application.isRequired,
  deleteApplication: PropTypes.func.isRequired,
  link: PropTypes.shape({ skip_payload_crypto: PropTypes.bool }),
  mayPurge: PropTypes.bool.isRequired,
  mayViewLink: PropTypes.bool.isRequired,
  onDeleteSuccess: PropTypes.func.isRequired,
  shouldConfirmDelete: PropTypes.bool.isRequired,
  updateApplication: PropTypes.func.isRequired,
  updateApplicationLink: PropTypes.func.isRequired,
}

ApplicationGeneralSettings.defaultProps = {
  link: {},
}

export default ApplicationGeneralSettings
