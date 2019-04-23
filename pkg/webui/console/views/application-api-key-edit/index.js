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
import { defineMessages } from 'react-intl'
import * as Yup from 'yup'
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
import diff from '../../../lib/diff'
import IntlHelmet from '../../../lib/components/intl-helmet'
import toast from '../../../components/toast'
import SubmitBar from '../../../components/submit-bar'

import { getApplicationApiKeyPageData } from '../../store/actions/application'
import api from '../../api'

import style from './application-api-key-edit.styl'

const m = defineMessages({
  deleteKey: 'Delete Key',
  modalWarning:
    'Are you sure you want to delete the {keyName} API Key? Deleting an application API Key cannot be undone!',
  keyEdit: 'Edit API Key',
  updateSuccess: 'Successfully updated API Key',
  deleteSuccess: 'Successfully deleted API Key',
})

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong),
  rights: Yup.object().test(
    'rights',
    sharedMessages.validateRights,
    values => Object.values(values).reduce((acc, curr) => acc || curr, false)
  ),
})

@connect(function ({ apiKeys, rights }, props) {
  const { appId, apiKeyId } = props.match.params

  const keysFetching = apiKeys.applications.fetching
  const rightsFetching = rights.applications.fetching
  const keysError = apiKeys.applications.error
  const rightsError = rights.applications.error

  const appKeys = apiKeys.applications[appId]
  const apiKey = appKeys ? appKeys.keys.find(k => k.id === apiKeyId) : undefined

  const appRights = rights.applications
  const rs = appRights ? appRights.rights : []

  return {
    keyId: apiKeyId,
    appId,
    apiKey,
    rights: rs,
    fetching: keysFetching || rightsFetching,
    error: keysError || rightsError,
  }
})
@withBreadcrumb('apps.single.api-keys.edit', function (props) {
  const { appId, keyId } = props

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/api-keys/${keyId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationApiKeyEdit extends React.Component {

  state = {
    error: '',
  }

  componentDidMount () {
    const { dispatch, appId } = this.props

    dispatch(getApplicationApiKeyPageData(appId))
  }

  async handleSubmit (values, { setSubmitting, resetForm }) {
    const { name, rights } = values
    const { appId, apiKey } = this.props

    const changed = diff({ name: apiKey.name }, { name })
    changed.rights = Object.keys(rights).filter(r => rights[r])

    await this.setState({ error: '' })

    try {
      await api.application.apiKeys.update(
        appId,
        apiKey.id,
        changed
      )
      resetForm({ ...values })
      toast({
        message: m.updateSuccess,
        type: toast.types.SUCCESS,
      })
    } catch (error) {
      resetForm({ ...values })
      await this.setState(error)
    }
  }

  async handleDelete () {
    const { dispatch, appId, keyId } = this.props

    await this.setState({ error: '' })

    try {
      await api.application.apiKeys.delete(appId, keyId)
      toast({
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
      dispatch(replace(`/console/applications/${appId}/api-keys`))
    } catch (error) {
      await this.setState(error)
    }
  }

  render () {
    const { apiKey, rights, fetching, error } = this.props

    if (error) {
      return 'ERROR'
    }

    if (fetching || !apiKey) {
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
        acc.rightsValues[right] = apiKey.rights.includes(right)

        return acc
      },
      {
        rightsItems: [],
        rightsValues: {},
      }
    )

    const initialFormValues = {
      id: apiKey.id,
      name: apiKey.name,
      rights: { ...rightsValues },
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <IntlHelmet title={m.keyEdit} />
            <Message component="h2" content={m.keyEdit} />
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
                title={sharedMessages.keyId}
                required
                valid
                disabled
                name="id"
                type="text"
              />
              <Field
                title={sharedMessages.name}
                name="name"
                type="text"
                autoFocus
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
                  message={m.deleteKey}
                  modalData={{
                    message: {
                      values: { keyName: apiKey.name ? `"${apiKey.name}"` : '' },
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
