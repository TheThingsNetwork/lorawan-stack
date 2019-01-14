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

import { withBreadcrumb } from '../../../components/breadcrumbs/context'
import Breadcrumb from '../../../components/breadcrumbs/breadcrumb'
import sharedMessages from '../../../lib/shared-messages'

import Form from '../../../components/form'
import Field from '../../../components/field'
import Button from '../../../components/button'
import Spinner from '../../../components/spinner'
import ModalButton from '../../../components/button/modal-button'
import Message from '../../../lib/components/message'

import { getApplicationsRightsList } from '../../../actions/applications'

import style from './application-access-edit.styl'

const m = defineMessages({
  deleteKey: 'Delete Key',
  modalWarning:
    'Are you sure you want to delete "{keyName}"? Deleting an application access key cannot be undone!',
})

const validationSchema = Yup.object().shape({
  name: Yup.string()
    .min(2, sharedMessages.validateTooShort)
    .max(50, sharedMessages.validateTooLong)
    .required(sharedMessages.validateRequired),
  description: Yup.string(),
})

@connect(function ({ apiKeys, rights }, props) {
  const { appId, apiKeyId } = props.match.params

  return {
    apiKeyId,
    appId,
    apiKey: apiKeys.applications[appId].keys.find(k => k.id === apiKeyId),
    applicationsRights: rights.applications.rights,
    fetching: rights.applications.fetching,
    error: rights.applications.error,
  }
})
@withBreadcrumb('apps.single.access.single', function (props) {
  const { appId, keyId } = props

  return (
    <Breadcrumb
      path={`/console/applications/${appId}/access/${keyId}/edit`}
      icon="general_settings"
      content={sharedMessages.edit}
    />
  )
})
@bind
export default class ApplicationAccessEdit extends React.Component {
  componentDidMount () {
    const { dispatch } = this.props

    dispatch(getApplicationsRightsList())
  }

  handleSubmit () { }

  handleDelete () { }

  handleCancel () { }

  render () {
    const { apiKey, applicationsRights, fetching, error } = this.props

    if (fetching ) {
      return <Spinner />
    }

    if (error) {
      return 'ERROR'
    }

    const { rightsItems, rightsValues } = applicationsRights.reduce(
      function (acc, right) {
        acc.rightsItems.push(
          <Field
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
      name: apiKey.name,
      description: apiKey.description,
      ...rightsValues,
    }

    return (
      <Container>
        <Row>
          <Col lg={8} md={12}>
            <Message component="h2" content={sharedMessages.edit} />
          </Col>
        </Row>
        <Row>
          <Col lg={8} md={12}>
            <Form
              onReset={this.handleCancel}
              horizontal
              onSubmit={this.handleSubmit}
              initialValues={initialFormValues}
              validationSchema={validationSchema}
            >
              <Message
                component="h4"
                content={sharedMessages.generalInformation}
              />
              <Field
                title={sharedMessages.name}
                required
                name="name"
                type="text"
              />
              <Field
                title={sharedMessages.description}
                name="description"
                type="text"
              />
              <Message component="h4" content={sharedMessages.rights} />
              {rightsItems}
              <div className={style.submitBar}>
                <div>
                  <Button type="submit" message={sharedMessages.saveChanges} />
                  <Button type="reset" naked secondary message={sharedMessages.cancel} />
                </div>
                <ModalButton
                  type="button"
                  icon="delete"
                  danger
                  naked
                  message={m.deleteKey}
                  modalData={{
                    message: {
                      values: { keyName: apiKey.name },
                      ...m.modalWarning,
                    },
                  }}
                  onApprove={this.handleDelete}
                />
              </div>
            </Form>
          </Col>
        </Row>
      </Container>
    )
  }
}
