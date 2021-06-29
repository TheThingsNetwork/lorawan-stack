// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { isEqual } from 'lodash'

import DeleteModalButton from '@ttn-lw/console/components/delete-modal-button'

import PageTitle from '@ttn-lw/components/page-title'
import { withBreadcrumb } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Form from '@ttn-lw/components/form'
import Input from '@ttn-lw/components/input'
import SubmitButton from '@ttn-lw/components/submit-button'
import toast from '@ttn-lw/components/toast'
import SubmitBar from '@ttn-lw/components/submit-bar'
import KeyValueMap from '@ttn-lw/components/key-value-map'

import withRequest from '@ttn-lw/lib/components/with-request'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'
import Require from '@console/lib/components/require'

import Yup from '@ttn-lw/lib/yup'
import diff from '@ttn-lw/lib/diff'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  checkFromState,
  mayEditBasicApplicationInfo,
  mayDeleteApplication,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayPurgeEntities,
} from '@console/lib/feature-checks'
import { attributeValidCheck, attributeTooShortCheck } from '@console/lib/attributes'

import { updateApplication, deleteApplication } from '@console/store/actions/applications'
import { getCollaboratorsList } from '@console/store/actions/collaborators'
import { getApiKeysList } from '@console/store/actions/api-keys'
import { getPubsubsList } from '@console/store/actions/pubsubs'
import { getWebhooksList } from '@console/store/actions/webhooks'

import {
  selectWebhooksTotalCount,
  selectWebhooksFetching,
  selectWebhooksError,
} from '@console/store/selectors/webhooks'
import {
  selectPubsubsTotalCount,
  selectPubsubsFetching,
  selectPubsubsError,
} from '@console/store/selectors/pubsubs'
import {
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorsError,
} from '@console/store/selectors/collaborators'
import {
  selectApiKeysTotalCount,
  selectApiKeysFetching,
  selectApiKeysError,
} from '@console/store/selectors/api-keys'
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
  state => {
    const mayViewApiKeys = checkFromState(mayViewOrEditApplicationApiKeys, state)
    const mayViewCollaborators = checkFromState(mayViewOrEditApplicationCollaborators, state)
    const apiKeysCount = selectApiKeysTotalCount(state)
    const collaboratorsCount = selectCollaboratorsTotalCount(state)
    const webhooksCount = selectWebhooksTotalCount(state)
    const pubsubsCount = selectPubsubsTotalCount(state)
    const mayPurgeApp = checkFromState(mayPurgeEntities, state)
    const mayDeleteApp = checkFromState(mayDeleteApplication, state)

    const entitiesFetching =
      selectApiKeysFetching(state) ||
      selectCollaboratorsFetching(state) ||
      selectPubsubsFetching(state) ||
      selectWebhooksFetching(state)
    const error =
      selectApiKeysError(state) ||
      selectCollaboratorsError(state) ||
      selectPubsubsError(state) ||
      selectWebhooksError(state)

    const fetching =
      entitiesFetching ||
      (mayViewApiKeys && typeof apiKeysCount === undefined) ||
      (mayViewCollaborators && collaboratorsCount === undefined) ||
      typeof collaboratorsCount === undefined ||
      typeof pubsubsCount === undefined
    const hasIntegrations = webhooksCount > 0 || pubsubsCount > 0
    const hasApiKeys = apiKeysCount > 0
    // Note: there is always at least one collaborator.
    const hasAddedCollaborators = collaboratorsCount > 1
    const isPristine = !hasApiKeys && !hasAddedCollaborators && !hasIntegrations
    return {
      appId: selectSelectedApplicationId(state),
      application: selectSelectedApplication(state),
      mayViewApiKeys,
      mayViewCollaborators,
      fetching,
      mayPurge: mayPurgeApp,
      shouldConfirmDelete:
        !isPristine || !mayViewCollaborators || !mayViewApiKeys || Boolean(error),
      mayDeleteApplication: mayDeleteApp,
    }
  },
  dispatch => ({
    ...bindActionCreators(
      {
        updateApplication: attachPromise(updateApplication),
        deleteApplication: attachPromise(deleteApplication),
        getApiKeysList,
        getCollaboratorsList,
        getWebhooksList,
        getPubsubsList,
      },
      dispatch,
    ),
    onDeleteSuccess: () => dispatch(replace(`/applications`)),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    deleteApplication: (id, purge = false) => dispatchProps.deleteApplication(id, { purge }),
    loadData: () => {
      if (stateProps.mayDeleteApplication) {
        if (stateProps.mayViewApiKeys) {
          dispatchProps.getApiKeysList('application', stateProps.appId)
        }

        if (stateProps.mayViewCollaborators) {
          dispatchProps.getCollaboratorsList('application', stateProps.appId)
        }

        dispatchProps.getWebhooksList(stateProps.appId)
        dispatchProps.getPubsubsList(stateProps.appId)
      }
    },
  }),
)
@withFeatureRequirement(mayEditBasicApplicationInfo, {
  redirect: ({ appId }) => `/applications/${appId}`,
})
@withRequest(({ loadData }) => loadData(), ({ fetching }) => fetching)
@withBreadcrumb('apps.single.general-settings', props => {
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
    appId: PropTypes.string.isRequired,
    application: PropTypes.application.isRequired,
    deleteApplication: PropTypes.func.isRequired,
    match: PropTypes.match.isRequired,
    mayPurge: PropTypes.bool.isRequired,
    onDeleteSuccess: PropTypes.func.isRequired,
    shouldConfirmDelete: PropTypes.bool.isRequired,
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
    if (isEqual(application.attributes || {}, appValues.attributes)) {
      delete appValues.attributes
    }

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
  async handleDelete(shouldPurge) {
    const { deleteApplication, onDeleteSuccess } = this.props
    const { appId } = this.props.match.params

    await this.setState({ error: '' })

    try {
      await deleteApplication(appId, shouldPurge)
      onDeleteSuccess()
    } catch (error) {
      await this.setState({ error })
    }
  }

  render() {
    const { appId, application, shouldConfirmDelete, mayPurge } = this.props
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
                  <DeleteModalButton
                    message={m.deleteApp}
                    entityId={appId}
                    entityName={application.name}
                    onApprove={this.handleDelete}
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
}
