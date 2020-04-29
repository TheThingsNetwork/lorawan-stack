// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import classnames from 'classnames'
import { defineMessages } from 'react-intl'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './json-payload.styl'

const m = defineMessages({
  invalid: 'Invalid JSON',
})

const OBJECT_PLACEHOLDER = '{…}'
const ARRAY_PLACEHOLDER = '[…]'

const KeyValue = ({ entryKey, entryValue, showKey }) => {
  let renderValue
  if (Array.isArray(entryValue)) {
    renderValue = ARRAY_PLACEHOLDER
  } else if (typeof entryValue === 'object') {
    renderValue = OBJECT_PLACEHOLDER
  } else {
    renderValue = JSON.stringify(entryValue)
  }

  const valueCls = classnames(style.value, {
    [style.string]: typeof entryValue === 'string',
    [style.number]: typeof entryValue === 'number',
    [style.boolean]: typeof entryValue === 'boolean',
    [style.value]: typeof entryValue === 'object',
  })

  const entries = []

  if (showKey) {
    entries.push(
      <span key={entryKey} className={style.key}>
        {entryKey}
      </span>,
    )
  }

  entries.push(
    <span key={`${entryKey}-value`} className={valueCls}>
      {renderValue}
    </span>,
  )

  return entries
}

KeyValue.propTypes = {
  entryKey: PropTypes.string.isRequired,
  entryValue: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.bool,
    PropTypes.number,
    PropTypes.shape({}),
    PropTypes.array,
  ]).isRequired,
  showKey: PropTypes.bool,
}

KeyValue.defaultProps = {
  showKey: false,
}

const JSONPayload = React.memo(props => {
  const { data } = props

  try {
    JSON.stringify(data)
  } catch (e) {
    return <Message className={style.invalid} content={m.invalid} />
  }

  const dataArray = Array.isArray(data)
  const dataObject = !dataArray && data !== null && typeof data === 'object'

  const cls = classnames(style.jsonPayload, {
    [style.jsonPayloadArray]: dataArray,
    [style.jsonPayloadObject]: dataObject,
  })

  const content = Object.keys(data).map(key => (
    <KeyValue key={key} entryKey={key} entryValue={data[key]} showKey={dataObject} />
  ))

  return <div className={cls}>{content}</div>
})

JSONPayload.propTypes = {
  data: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.bool,
    PropTypes.number,
    PropTypes.shape({}),
    PropTypes.array,
  ]).isRequired,
}

JSONPayload.defaultProps = {}

export default JSONPayload
