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

import classNames from 'classnames'
import React from 'react'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './field.styl'

const FieldError = ({ content, error, warning, title, className, id }) => {
  const icon = error ? 'error' : 'warning'
  const contentValues = content.values || {}
  const classname = classNames(style.message, className, {
    [style.show]: content && content !== '',
    [style.hide]: !content || content === '',
    [style.err]: error,
    [style.warn]: warning,
  })

  if (title) {
    contentValues.field = <Message content={title} className={style.name} key={title.id || title} />
  }

  return (
    <div className={classname} id={id}>
      <Icon icon={icon} className={style.icon} />
      <Message content={content.message || content} values={contentValues} />
    </div>
  )
}

FieldError.propTypes = {
  className: PropTypes.string,
  content: PropTypes.oneOfType([
    PropTypes.error,
    PropTypes.shape({
      message: PropTypes.error.isRequired,
      values: PropTypes.shape({}).isRequired,
    }),
  ]).isRequired,
  error: PropTypes.bool,
  id: PropTypes.string.isRequired,
  title: PropTypes.message,
  warning: PropTypes.bool,
}

FieldError.defaultProps = {
  className: undefined,
  title: undefined,
  warning: false,
  error: false,
}

export default FieldError
