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

import React, { useCallback } from 'react'
import classNames from 'classnames'

import Button from '@ttn-lw/components/button-v2'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './search-button.styl'

const SearchButton = ({ onClick, className }) => {
  const handleClick = useCallback(() => {
    onClick()
  }, [onClick])
  return (
    <Button onClick={handleClick} naked className={classNames(style.searchButton, className)}>
      <div className="d-flex gap-cs-xxs">
        <Icon icon="search" className={style.icon} />
        <Message content={sharedMessages.search} component="p" className="m-0" />
      </div>
      <div className={style.backslashContainer}>
        <Message content="/" component="p" className={style.backslash} />
      </div>
    </Button>
  )
}

SearchButton.propTypes = {
  className: PropTypes.string,
  onClick: PropTypes.func.isRequired,
}

SearchButton.defaultProps = {
  className: undefined,
}

export default SearchButton
