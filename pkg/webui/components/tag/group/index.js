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
import bind from 'autobind-decorator'
import classnames from 'classnames'

import PropTypes from '../../../lib/prop-types'
import Tag from '../.'

import style from './group.styl'

const measureWidth = function(element) {
  if (!element) {
    return 0
  }

  return element.current.clientWidth
}

// the width in pixels for the left tag
const LEFT_TAG_WIDTH = 40
// space between the tags
const TAG_SPACE_WIDTH = 3

@bind
class TagGroup extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      left: 0,
    }

    this.element = React.createRef()
  }

  componentDidMount() {
    window.addEventListener('resize', this.handleWindowResize)

    this.handleWindowResize()
  }

  componentDidUpdate(prevProps) {
    const props = this.props

    // calculate fit on any props change
    if (prevProps.tags !== props.tags) {
      this.checkTagsFit()
    }
  }

  checkTagsFit() {
    this.handleWindowResize()
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleWindowResize)
  }

  handleWindowResize() {
    const { tags, tagMaxWidth } = this.props

    const containerWidth = measureWidth(this.element)
    const totalTagCount = tags.length
    const possibleFitCount = Math.floor(containerWidth / tagMaxWidth) || 1

    // count for the left tag and paddings between tags
    const leftTagWidth = totalTagCount !== possibleFitCount ? LEFT_TAG_WIDTH : 0
    const spaceWidth = possibleFitCount > 1 ? possibleFitCount * TAG_SPACE_WIDTH : 0

    const finalAvailableWidth = containerWidth - leftTagWidth - spaceWidth
    const finalLeft = Math.floor(finalAvailableWidth / tagMaxWidth) || 1

    this.setState({
      left: totalTagCount - finalLeft,
    })
  }

  render() {
    const { className, tags } = this.props
    const { left } = this.state

    const ts = tags.slice(0, tags.length - left)

    return (
      <div ref={this.element} className={classnames(className, style.group)}>
        {ts}
        {left > 0 && <Tag content={`+${left}`} />}
      </div>
    )
  }
}

TagGroup.propTypes = {
  tagMaxWidth: PropTypes.number.isRequired,
  tags: PropTypes.arrayOf(PropTypes.any).isRequired,
}

export default TagGroup
