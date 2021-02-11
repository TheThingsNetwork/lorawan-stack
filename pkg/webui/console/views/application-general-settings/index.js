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
import bind from 'autobind-decorator'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { connect } from 'react-redux'
import { replace } from 'connected-react-router'
import { bindActionCreators } from 'redux'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'
import SubmitBar from '@ttn-lw/components/submit-bar'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'
import Require from '@console/lib/components/require'

import Yup from '@ttn-lw/lib/yup'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { mayEditBasicApplicationInfo, mayDeleteApplication } from '@console/lib/feature-checks'
import { attributeValidCheck, attributeTooShortCheck } from '@console/lib/attributes'

import { updateApplication, deleteApplication } from '@console/store/actions/applications'

import {
  selectSelectedApplication,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

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
    ),
})

@connect(
  state => ({
    appId: selectSelectedApplicationId(state),
    application: selectSelectedApplication(state),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        updateApplication: attachPromise(updateApplication),
        deleteApplication: attachPromise(deleteApplication),
      },
      dispatch,
    ),
    onDeleteSuccess: () => dispatch(replace(`/applications`)),
  }),
)
@withFeatureRequirement(mayEditBasicApplicationInfo, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withBreadcrumb('apps.single.general-settings', function (props) {
  const { appId } = props

  return (
    <Breadcrumb
      path={`/applications/${appId}/general-settings`}
      content={sharedMessages.generalSettings}
    />
  )
})
export default class ApplicationGeneralSettings extends React.Component {
  static propTypes = {
    application: PropTypes.application.isRequired,
    deleteApplication: PropTypes.func.isRequired,
    match: PropTypes.match.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    updateApplication: PropTypes.func.isRequired,
  }

  state = {
    error: '',
  }

  @bind
  async handleSubmit(values, { resetForm, setSubmitting }) {
    const { application, updateApplication } = this.props

    await this.setState({ error: '' })

    const appValues = mapFormValuesToApplication(values)

    const changed = diff(application, appValues)

    // If there is a change in attributes, copy all attributes so they don't get
    // overwritten.
    const update =
      'attributes' in changed ? { ...changed, attributes: appValues.attributes } : changed

    const {
      ids: { application_id },
    } = application

    try {
      await updateApplication(application_id, update)
      resetForm({ values })
      toast({
        title: application_id,
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      setSubmitting(false)
      await this.setState({ error })
    }
  }

  @bind
  async handleDelete() {
    const { deleteApplication, onDeleteSuccess } = this.props
    const { appId } = this.props.match.params

    await this.setState({ error: '' })

    try {
      await deleteApplication(appId)
      onDeleteSuccess()
    } catch (error) {
      await this.setState({ error })
    }
  }

  render() {
    const { application } = this.props
    const { error } = this.state
    const initialValues = mapApplicationToFormValues(application)

    return (
      <Container>
        <PageTitle title={sharedMessages.generalSettings} />
        <Row>
          <Col lg={8} md={12}>
            <Form
              error={error}
              onSubmit={this.handleSubmit}
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
              <SubmitBar>
                <Form.Submit component={SubmitButton} message={sharedMessages.saveChanges} />
                <Require featureCheck={mayDeleteApplication}>
                  <ModalButton
                    type="button"
                    icon="delete"
                    danger
                    naked
                    message={m.deleteApp}
                    modalData={{
                      message: {
                        values: {
                          appName: application.name
                            ? application.name
                            : application.ids.application_id,
                        },
                        ...m.modalWarning,
                      },
                    }}
                    onApprove={this.handleDelete}
                  />
                </Require>
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
