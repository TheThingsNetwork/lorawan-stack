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

import React from 'react'
import { connect } from 'react-redux'
import { defineMessages, useIntl } from 'react-intl'
import { Col, Row, Container } from 'react-grid-system'
import { bindActionCreators } from 'redux'
import { push } from 'connected-react-router'

import toast from '@ttn-lw/components/toast'
import Button from '@ttn-lw/components/button'
import DataSheet from '@ttn-lw/components/data-sheet'
import Tag from '@ttn-lw/components/tag'
import TagGroup from '@ttn-lw/components/tag/group'

import DateTime from '@ttn-lw/lib/components/date-time'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { deleteAuthorization } from '@account/store/actions/authorizations'

import { selectSelectedAuthorization } from '@account/store/selectors/authorizations'

const RIGHT_TAG_MAX_WIDTH = 140

const m = defineMessages({
  deleteButton: 'De-authorize this client',
  deleteSuccess: 'This oauth client was successfully de-authorized',
  deleteFailure: 'There was an error and this client could not be de-authorized',
})

const AuthorizationSettings = props => {
  const {
    authorization: { client_ids, user_ids, rights, created_at, updated_at },
    deleteAuthorization,
    navigateToList,
  } = props
  const { client_id } = client_ids
  const { user_id } = user_ids
  const intl = useIntl()

  const handleDeleteAuthorization = React.useCallback(async () => {
    try {
      await deleteAuthorization(user_id, client_id)
      toast({
        title: client_id,
        message: m.deleteSuccess,
        type: toast.types.SUCCESS,
      })
      navigateToList()
    } catch (err) {
      toast({
        title: client_id,
        message: m.deleteFailure,
        type: toast.types.ERROR,
      })
    }
  }, [navigateToList, deleteAuthorization, client_id, user_id])

  const tags = rights.map(r => {
    let rightLabel = intl.formatMessage({ id: `enum:${r}` })
    rightLabel = rightLabel.charAt(0).toUpperCase() + rightLabel.slice(1)
    return <Tag content={rightLabel} key={r} />
  })

  const sheetData = React.useMemo(
    () => [
      {
        header: sharedMessages.generalInformation,
        items: [
          {
            key: sharedMessages.oauthClientId,
            value: user_id,
            type: 'code',
            sensitive: false,
          },
          { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
          { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
          {
            key: sharedMessages.rights,
            value: <TagGroup tagMaxWidth={RIGHT_TAG_MAX_WIDTH} tags={tags} />,
            sensitive: false,
          },
        ],
      },
      {
        header: sharedMessages.actions,
        items: [
          {
            value: (
              <Button
                type="button"
                onClick={handleDeleteAuthorization}
                message={m.deleteButton}
                icon="delete"
                danger
              />
            ),
          },
        ],
      },
    ],
    [created_at, user_id, updated_at, tags, handleDeleteAuthorization],
  )

  return (
    <>
      <Container>
        <Row>
          <Col sm={12} lg={6}>
            <DataSheet data={sheetData} />
          </Col>
        </Row>
      </Container>
    </>
  )
}

AuthorizationSettings.propTypes = {
  authorization: PropTypes.shape({
    client_ids: PropTypes.shape({
      client_id: PropTypes.string,
    }),
    user_ids: PropTypes.shape({
      user_id: PropTypes.string,
    }),
    rights: PropTypes.rights,
    created_at: PropTypes.string,
    updated_at: PropTypes.string,
  }).isRequired,
  deleteAuthorization: PropTypes.func.isRequired,
  navigateToList: PropTypes.func.isRequired,
}

export default connect(
  (state, props) => ({
    authorization: selectSelectedAuthorization(state, props.match.params.clientId),
  }),
  dispatch => ({
    ...bindActionCreators(
      {
        deleteAuthorization: attachPromise(deleteAuthorization),
      },
      dispatch,
    ),
    navigateToList: () => dispatch(push('/client-authorizations')),
  }),
  (stateProps, dispatchProps, ownProps) => ({
    ...stateProps,
    ...dispatchProps,
    ...ownProps,
    deleteAuthorization: (userId, clientId) => dispatchProps.deleteAuthorization(userId, clientId),
  }),
)(AuthorizationSettings)
