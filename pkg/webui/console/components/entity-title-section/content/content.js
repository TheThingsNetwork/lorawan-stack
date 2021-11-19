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

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './content.styl'

const EntityTitleSectionContent = props => {
  const { className, children, bottomBarLeft, bottomBarRight, fetching } = props

  return (
    <div className={classnames(className, style.container)}>
      {fetching ? (
        <Spinner after={0} faded micro inline>
          <Message content={sharedMessages.fetching} />
        </Spinner>
      ) : (
        <div className={style.content}>
          {Boolean(bottomBarLeft || bottomBarRight) && (
            <div className={style.bottomBar}>
              <div>{bottomBarLeft}</div>
              <div>{bottomBarRight}</div>
            </div>
          )}
          {children}
        </div>
      )}
    </div>
  )
}

EntityTitleSectionContent.propTypes = {
  bottomBarLeft: PropTypes.node,
  bottomBarRight: PropTypes.node,
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  className: PropTypes.string,
  fetching: PropTypes.bool,
}

EntityTitleSectionContent.defaultProps = {
  bottomBarLeft: undefined,
  bottomBarRight: undefined,
  className: undefined,
  children: undefined,
  fetching: false,
}

export default EntityTitleSectionContent
