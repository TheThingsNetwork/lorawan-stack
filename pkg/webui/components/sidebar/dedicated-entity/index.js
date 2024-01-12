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
import { Link } from 'react-router-dom'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './dedicated-entity.styl'

const DedicatedEntity = ({ label, className, buttonMessage, path, backPath }) => (
  <div className={classnames(className, style.dedicatedEntity)}>
    <Link to={backPath} className={style.dedicatedEntityCurtain}>
      <Button
        className={style.dedicatedEntityButton}
        icon="arrow_left_alt"
        message={buttonMessage}
      />
    </Link>
    <Link className={classnames(style.dedicatedEntityItem, 'd-flex', 'al-center')} to={path}>
      <hr className={style.dedicatedEntityDivider} />
      <Message content={label} className={style.dedicatedEntityLabel} component="p" />
    </Link>
  </div>
)

DedicatedEntity.propTypes = {
  backPath: PropTypes.string.isRequired,
  buttonMessage: PropTypes.message.isRequired,
  className: PropTypes.string,
  label: PropTypes.string.isRequired,
  path: PropTypes.string.isRequired,
}

DedicatedEntity.defaultProps = {
  className: undefined,
}

export default DedicatedEntity
