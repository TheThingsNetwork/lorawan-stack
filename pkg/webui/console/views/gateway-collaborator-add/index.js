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
import Field from '../../../components/field'
import Button from '../../../components/button'
import Message from '../../../lib/components/message'
import FieldGroup from '../../../components/field/group'
import IntlHelmet from '../../../lib/components/intl-helmet'
import { id as collaboratorIdRegexp } from '../../lib/regexp'
import SubmitBar from '../../../components/submit-bar'

import { getGatewaysRightsList } from '../../store/actions/gateways'
import api from '../../api'

import style from './gateway-collaborator-add.styl'

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

@connect(function ({ rights, collaborators }, props) {
  const gtwId = props.match.params.gtwId

  return {
    gtwId,
    collaborators: collaborators.gateways.collaborators,
    fetching: rights.gateways.fetching,
    error: rights.gateways.error,
    rights: rights.gateways.rights,
  }
})
@withBreadcrumb('gtws.single.collaborators.add', function (props) {
  const gtwId = props.gtwId
  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/collaborators/add`}
      icon="add"
      content={sharedMessages.add}
    />
  )
})
@bind
export default class GatewayCollaboratorAdd extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { dispatch, gtwId } = this.props

    dispatch(getGatewaysRightsList(gtwId))
  }

  async handleSubmit (values, { resetForm }) {
    const { collaborator_id, collaborator_type, rights } = values
    const { gtwId, dispatch } = this.props

    const collaborator_ids = {
      [`${collaborator_type}_ids`]: {
        [`${collaborator_type}_id`]: collaborator_id,
      },
    }
    const collaborator = {
      ids: collaborator_ids,
      rights: Object.keys(rights).filter(r => rights[r]),
    }

    try {
      await api.gateway.collaborators.add(gtwId, collaborator)
      dispatch(push(`/console/gateways/${gtwId}/collaborators`))
    } catch (error) {
      resetForm(values)
      this.setState({ error })
    }
  }

  render () {
    const { rights, fetching, error } = this.props

    if (error) {
      throw error
    }

    if (fetching && !rights.length) {
      return <Spinner center />
    }

    const { rightsItems, rightsValues } = rights.reduce(
      function (acc, right) {
        acc.rightsItems.push(
          <Field
            className={style.rightLabel}
            key={right}
            name={right}
            type="checkbox"
            title={{ id: `enum:${right}` }}
            form
          />
        )
        acc.rightsValues[right] = false

        return acc
      },
      {
        rightsItems: [],
        rightsValues: {},
      }
    )

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
              mapErrorsToFields={{
                user_not_found: 'collaborator_id',
                organization_not_found: 'collaborator_id',
              }}
            >
              <Message
                component="h4"
                content={sharedMessages.generalInformation}
              />
              <Field
                name="collaborator_id"
                type="text"
                title={sharedMessages.collaboratorId}
                required
                autoFocus
              />
              <Field
                type="select"
                name="collaborator_type"
                title={sharedMessages.type}
                required
                options={[
                  { value: 'user', label: sharedMessages.user },
                  { value: 'organization', label: sharedMessages.organization },
                ]}
              />
              <FieldGroup
                name="rights"
                title={sharedMessages.rights}
              >
                {rightsItems}
              </FieldGroup>
              <SubmitBar>
                <Button type="submit" message={sharedMessages.addCollaborator} />
              </SubmitBar>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
