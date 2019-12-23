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
import { defineMessages } from 'react-intl'
import bind from 'autobind-decorator'

import Button from '../../button'

import PropTypes from '../../../lib/prop-types'
import style from './details.styl'

const m = defineMessages({
  showDetails: 'Show Details',
  hideDetails: 'Hide Details',
})

@bind
export default class Details extends React.PureComponent {
  static propTypes = {
    details: PropTypes.oneOfType([PropTypes.string, PropTypes.error]).isRequired,
  }

  state = {
    expanded: false,
    buttonIcon: 'arrow_drop_down',
    buttonText: m.showDetails,
  }

  showDropdown() {
    document.addEventListener('mousedown', this.handleClickOutside)
    this.setState({
      expanded: true,
      buttonIcon: 'arrow_drop_up',
      buttonText: m.hideDetails,
    })
  }

  hideDropdown() {
    document.removeEventListener('mousedown', this.handleClickOutside)
    this.setState({
      expanded: false,
      buttonIcon: 'arrow_drop_down',
      buttonText: m.showDetails,
    })
  }

  toggleDropdown() {
    let { expanded } = this.state
    expanded = !expanded
    if (expanded) {
      this.showDropdown()
    } else {
      this.hideDropdown()
    }
  }

  render() {
    const { expanded, buttonIcon, buttonText } = this.state
    const { details } = this.props
    const content = typeof details === 'string' ? details : JSON.stringify(details, undefined, 2)
    return (
      <div className={style.details}>
        <Button
          className={style.detailsButton}
          naked
          secondary
          onClick={this.toggleDropdown}
          icon={buttonIcon}
          message={buttonText}
          type="button"
        />
        {expanded && <pre className={style.detailsDropdown}>{content}</pre>}
      </div>
    )
  }
}

export { Details }
