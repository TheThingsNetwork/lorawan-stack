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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { Container, Col, Row } from 'react-grid-system'
import * as Yup from 'yup'
import { defineMessages } from 'react-intl'
import { replace } from 'connected-react-router'

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'
import Form from '../../../components/form'
import SubmitButton from '../../../components/submit-button'
import Input from '../../../components/input'
import Spinner from '../../../components/spinner'
import ModalButton from '../../../components/button/modal-button'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'
import SubmitBar from '../../../components/submit-bar'
import RightsGroup from '../../components/rights-group'

import { getApplicationCollaboratorsList } from '../../store/actions/application'
import { getApplicationsRightsList } from '../../store/actions/applications'
import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '../../store/selectors/applications'

import api from '../../api'

const validationSchema = Yup.object().shape({
  rights: Yup.object().test(
    'rights',
    sharedMessages.validateRights,
    values => Object.values(values).reduce((acc, curr) => acc || curr, false)
  ),
})

const m = defineMessages({
  deleteSuccess: 'Successfully removed collaborator',
  updateSuccess: 'Successfully updated collaborator rights',
  modalWarning:
    'Are you sure you want to remove {collaboratorId} as a collaborator?',
})

@connect(function (state, props) {
  const appId = selectSelectedApplicationId(state)
  const { collaboratorId } = props.match.params
  const collaboratorsFetching = state.collaborators.applications.fetching
  const collaboratorsError = state.collaborators.applications.error

  const appCollaborators = state.collaborators.applications[appId]
  const collaborator = appCollaborators ? appCollaborators.collaborators
    .find(c => c.id === collaboratorId) : undefined

  const fetching = selectApplicationRightsFetching(state) || collaboratorsFetching
  const error = selectApplicationRightsError(state) || collaboratorsError

  return {
    collaboratorId,
    collaborator,
    appId,
    rights: selectApplicationRights(state),
    universalRights: selectApplicationUniversalRights(state),
    fetching,
    error,
  }
}, dispatch => ({
  async loadData (appId) {
    await dispatch(getApplicationsRightsList(appId))
    dispatch(getApplicationCollaboratorsList(appId))
  },
  redirectToList (appId) {
    dispatch(replace(`/console/applications/${appId}/collaborators`))
  },
}))
@withBreadcrumb('apps.single.collaborators.edit', function (props) {
  const { appId, collaboratorId } = props

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/collaborators/${collaboratorId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationCollaboratorEdit extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { loadData, appId } = this.props

    loadData(appId)
  }

  async handleSubmit (values, { resetForm }) {
    const { collaborator_id, rights } = values
    const { appId, collaborator } = this.props
    const collaborator_type = collaborator.isUser ? 'user' : 'organization'

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
      rights: Object.keys(rights).filter(r => rights[r]),
    }

    await this.setState({ error: '' })

    try {
      await api.application.collaborators.update(appId, updatedCollaborator)
      resetForm(values)
      toast({
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      resetForm(values)
      await this.setState({ error })
    }
  }

  async handleDelete () {
    const { collaborator, redirectToList, appId } = this.props
    const collaborator_type = collaborator.isUser ? 'user' : 'organization'

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator.id,
      },
    }
    const updatedCollaborator = {
      ids: collaborator_ids,
    }

    try {
      await api.application.collaborators.remove(appId, updatedCollaborator)
      toast({
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
      redirectToList(appId)
    } catch (error) {
      await this.setState({ error })
    }
  }

  render () {
    const { collaborator, rights, fetching, error, universalRights } = this.props

    if (error) {
      throw error
    }

    if (fetching || !collaborator) {
      return <Spinner center />
    }

    const hasUniversalRights = universalRights.reduce(
      (acc, curr) => acc || collaborator.rights.includes(curr), false)
    const rightsValues = rights.reduce(
      function (acc, right) {
        acc[right] = hasUniversalRights || collaborator.rights.includes(right)

        return acc
      }, {})

    const initialFormValues = {
      collaborator_id: collaborator.id,
      rights: { ...rightsValues },
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet
              title={sharedMessages.collaboratorEdit}
              values={{ collaboratorId: collaborator.id }}
            />
            <Message
              component="h2"
              content={sharedMessages.collaboratorEditRights}
              values={{ collaboratorId: collaborator.id }}
            />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <Form
              horizontal
              error={this.state.error}
              onSubmit={this.handleSubmit}
              initialValues={initialFormValues}
              validationSchema={validationSchema}
            >
              <Message
                component="h4"
                content={sharedMessages.generalInformation}
              />
              <Form.Field
                title={sharedMessages.collaboratorId}
                required
                valid
                disabled
                name="collaborator_id"
                component={Input}
              />
              <Form.Field
                name="rights"
                title={sharedMessages.rights}
                required
                component={RightsGroup}
                rights={rights}
                universalRight={universalRights[0]}
              />
              <SubmitBar>
                <Form.Submit
                  component={SubmitButton} message={sharedMessages.saveChanges}
                />
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={sharedMessages.removeCollaborator}
                  modalData={{
                    message: {
                      values: { collaboratorId: collaborator.id },
                      ...m.modalWarning,
                    },
                  }}
                  onApprove={this.handleDelete}
                />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
