// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

const ContentTemplate = ({ messages, values, withList, listTitle, listElement }) => (
  <>
    <Message content={messages.body} values={{ ...values.body }} component="p" />
    {'entities' in messages && (
      <Message content={messages.entities} values={{ ...values.entities }} component="p" />
    )}
    {withList && (
      <>
        <p>
          <Message component="b" content={listTitle} />
        </p>
        <ul>
          {listElement.map(el => (
            <>
              <Message component="li" content={el} />
              <Message content={{ id: `enum:${el}` }} firstToUpper />
            </>
          ))}
        </ul>
      </>
    )}
    <Message content={messages.action} values={{ ...values.action }} />
  </>
)

ContentTemplate.propTypes = {
  listElement: PropTypes.arrayOf(PropTypes.string),
  listTitle: PropTypes.message,
  messages: PropTypes.shape({
    body: PropTypes.message.isRequired,
    entities: PropTypes.message,
    action: PropTypes.message.isRequired,
  }).isRequired,
  values: PropTypes.shape({
    body: PropTypes.shape({}).isRequired,
    entities: PropTypes.shape({}),
    action: PropTypes.shape({}).isRequired,
  }).isRequired,
  withList: PropTypes.bool,
}

ContentTemplate.defaultProps = {
  withList: false,
  listTitle: undefined,
  listElement: undefined,
}

export default ContentTemplate
