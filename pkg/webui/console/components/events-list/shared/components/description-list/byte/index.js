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

import SafeInspector from '@ttn-lw/components/safe-inspector'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import DescriptionListItem from '../item'

import style from './byte.styl'

const DescriptionListByteItem = props => {
  const { title, data } = props

  const isEmpty = data.length === 0

  return (
    <DescriptionListItem title={title}>
      {isEmpty ? (
        <Message className={style.empty} firstToLower content={sharedMessages.empty} />
      ) : (
        <SafeInspector
          data={data}
          hideable={false}
          initiallyVisible
          small
          noCopy
          noTransform
          disableResize
        />
      )}
    </DescriptionListItem>
  )
}

DescriptionListByteItem.propTypes = {
  data: PropTypes.string,
  title: PropTypes.message,
}

DescriptionListByteItem.defaultProps = {
  data: '',
  title: undefined,
}

export default DescriptionListByteItem
