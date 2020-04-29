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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './column.styl'

const HeaderColumn = props => {
  const { className, content } = props

  return (
    <div className={classnames(className, style.column)}>
      <Message className={style.content} content={content} firstToUpper />
    </div>
  )
}

HeaderColumn.propTypes = {
  className: PropTypes.string,
  content: PropTypes.message.isRequired,
}

HeaderColumn.defaultProps = {
  className: undefined,
}

export default HeaderColumn
