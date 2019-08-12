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
import classnames from 'classnames'
import { defineMessages } from 'react-intl'
import { Link } from 'react-router-dom'
import PropTypes from '../../lib/prop-types'

import Message from '../../lib/components/message'

import style from './footer.styl'

const m = defineMessages({
  footer: "You are the network. Let's build this thing together.",
})

const Footer = function({ className, links = [] }) {
  return (
    <footer className={classnames(className, style.footer)}>
      <div>
        <span>
          <Message content={m.footer} /> –{' '}
        </span>
        <a className={style.link} href="https://www.thethingsnetwork.org">
          The Things Network
        </a>
      </div>
      <div>
        {links.map((item, key) => (
          <Link key={key} className={style.link} to={item.link}>
            <Message content={item.title} />
          </Link>
        ))}
        <span className={style.version}>v{process.env.VERSION}</span>
      </div>
    </footer>
  )
}

Footer.propTypes = {
  /**
   * A list of links to be displayed in the footer component
   * @param {(string|Object)} title - The title of the link
   * @param {string} link - The link url
   */
  links: PropTypes.arrayOf(
    PropTypes.shape({
      title: PropTypes.message.isRequired,
      link: PropTypes.string.isRequired,
    }),
  ),
}

export default Footer
