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
import { Container, Col, Row } from 'react-grid-system'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import * as Yup from 'yup'
import { push } from 'connected-react-router'

import Spinner from '../../../components/spinner'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import sharedMessages from '../../../lib/shared-messages'
import Form from '../../../components/form'
import SubmitButton from '../../../components/submit-button'
import Input from '../../../components/input'
import Select from '../../../components/select'
import Message from '../../../lib/components/message'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { id as collaboratorIdRegexp } from '../../lib/regexp'
import SubmitBar from '../../../components/submit-bar'
import RightsGroup from '../../components/rights-group'

import { getApplicationsRightsList } from '../../store/actions/applications'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import {
  selectApplicationRights,
  selectApplicationUniversalRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '../../store/selectors/application'

import api from '../../api'

const validationSchema = Yup.object().shape({
  collaborator_id: Yup.string()
    .matches(collaboratorIdRegexp, sharedMessages.validateAlphanum)
    .required(sharedMessages.validateRequired),
  collaborator_type: Yup.string()
    .required(sharedMessages.validateRequired),
  rights: Yup.object().test(
    'rights',
    sharedMessages.validateRights,
    values => Object.values(values).reduce((acc, curr) => acc || curr, false)
  ),
})

@connect(function (state) {
  return {
    appId: selectSelectedApplicationId(state),
    collaborators: state.collaborators.applications.collaborators,
    rights: selectApplicationRights(state),
    universalRights: selectApplicationUniversalRights(state),
    fetching: selectApplicationRightsFetching(state),
    error: selectApplicationRightsError(state),
  }
})
@withBreadcrumb('apps.single.collaborators.add', function (props) {
  const appId = props.appId
  return (
    <Breadcrumb
      path={`/console/applications/${appId}/collaborators/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class ApplicationCollaboratorAdd extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { dispatch, appId } = this.props

    dispatch(getApplicationsRightsList(appId))
  }

  async handleSubmit (values, { resetForm }) {
    const { collaborator_id, collaborator_type, rights } = values
    const { appId, dispatch } = this.props

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }
    const collaborator = {
      ids: collaborator_ids,
      rights: Object.keys(rights).filter(r => rights[r]),
    }

    await this.setState({ error: '' })

    try {
      await api.application.collaborators.add(appId, collaborator)
      dispatch(push(`/console/applications/${appId}/collaborators`))
    } catch (error) {
      resetForm(values)
      this.setState({ error })
    }
  }

  render () {
    const { rights, fetching, error, universalRights } = this.props

    if (error) {
      throw error
    }

    if (fetching && !rights.length) {
      return <Spinner center />
    }

    const rightsValues = rights.reduce(function (acc, right) {
      acc[right] = false

      return acc
    }, {})

    const initialFormValues = {
      collaborator_id: '',
      collaborator_type: 'user',
      rights: rightsValues,
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={sharedMessages.addCollaborator} />
            <Message component="h2" content={sharedMessages.addCollaborator} />
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
                name="collaborator_id"
                component={Input}
                title={sharedMessages.collaboratorId}
                required
                autoFocus
              />
              <Form.Field
                name="collaborator_type"
                component={Select}
                title={sharedMessages.type}
                required
                options={[
                  { value: 'user', label: sharedMessages.user },
                  { value: 'organization', label: sharedMessages.organization },
                ]}
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
                  component={SubmitButton}
                  message={sharedMessages.addCollaborator}
                />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
