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
import Field from '../../../components/field'
import Button from '../../../components/button'
import Spinner from '../../../components/spinner'
import ModalButton from '../../../components/button/modal-button'
import Message from '../../../lib/components/message'
import FieldGroup from '../../../components/field/group'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'
import SubmitBar from '../../../components/submit-bar'

import { getGatewayCollaboratorsList } from '../../store/actions/gateway'
import { getGatewaysRightsList } from '../../store/actions/gateways'
import api from '../../api'

import style from './gateway-collaborator-edit.styl'

// TODO: Move this to checkbox group later, see https://github.com/TheThingsNetwork/lorawan-stack/issues/189
const UNIVERSAL_APPLICATION_RIGHTS = [ 'RIGHT_APPLICATION_ALL', 'RIGHT_ALL' ]

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

@connect(function ({ collaborators, rights }, props) {
  const { gtwId, collaboratorId } = props.match.params
  const collaboratorsFetching = collaborators.gateways.fetching
  const rightsFetching = rights.gateways.fetching
  const collaboratorsError = collaborators.gateways.error
  const rightsError = rights.gateways.error

  const gtwRights = rights.gateways
  const rs = gtwRights ? gtwRights.rights : []

  const gtwCollaborators = collaborators.gateways[gtwId]
  const collaborator = gtwCollaborators ? gtwCollaborators.collaborators
    .find(c => c.id === collaboratorId) : undefined

  return {
    collaboratorId,
    collaborator,
    gtwId,
    rights: rs,
    fetching: collaboratorsFetching || rightsFetching,
    error: collaboratorsError || rightsError,
  }
}, dispatch => ({
  loadData (gtwId) {
    dispatch(getGatewaysRightsList(gtwId))
    dispatch(getGatewayCollaboratorsList(gtwId))
  },
  redirectToList (gtwId) {
    dispatch(replace(`/console/gateways/${gtwId}/collaborators`))
  },
}))
@withBreadcrumb('gtws.single.collaborators.edit', function (props) {
  const { gtwId, collaboratorId } = props

  return (
    <Breadcrumb
      path={`/console/gateways/${gtwId}/collaborators/${collaboratorId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class GatewayCollaboratorEdit extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { loadData, gtwId } = this.props

    loadData(gtwId)
  }

  async handleSubmit (values, { resetForm }) {
    const { collaborator_id, rights } = values
    const { gtwId, collaborator } = this.props
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
      await api.gateway.collaborators.update(gtwId, updatedCollaborator)
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
    const { collaborator, redirectToList, gtwId } = this.props
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
      await api.gateway.collaborators.remove(gtwId, updatedCollaborator)
      toast({
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
      redirectToList(gtwId)
    } catch (error) {
      await this.setState({ error })
    }
  }

  render () {
    const { collaborator, rights, fetching, error } = this.props

    if (error) {
      throw error
    }

    if (fetching || !collaborator) {
      return <Spinner center />
    }

    const hasUniversalRights = UNIVERSAL_APPLICATION_RIGHTS.reduce(
      (acc, curr) => acc || collaborator.rights.includes(curr), false)
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
        acc.rightsValues[right] = hasUniversalRights || collaborator.rights.includes(right)

        return acc
      },
      {
        rightsItems: [],
        rightsValues: {},
      }
    )

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
              <Field
                title={sharedMessages.collaboratorId}
                required
                valid
                disabled
                name="collaborator_id"
                type="text"
              />
              <FieldGroup
                name="rights"
                title={sharedMessages.rights}
              >
                {rightsItems}
              </FieldGroup>
              <SubmitBar>
                <Button type="submit" message={sharedMessages.saveChanges} />
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
