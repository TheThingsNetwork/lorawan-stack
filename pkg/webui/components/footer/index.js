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
import Button from '../../components/button'
import Offline from '../offline'

import style from './footer.styl'

const m = defineMessages({
  footer: "You are the network. Let's build this thing together.",
  getSupport: 'Get Support',
})

const Footer = function({ className, links, supportLink }) {
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
        <Offline />
        <span className={style.version}>v{process.env.VERSION}</span>
        {supportLink && (
          <Button.AnchorLink
            message={m.getSupport}
            icon="contact_support"
            href={supportLink}
            target="_blank"
          />
        )}
      </div>
    </footer>
  )
}

Footer.propTypes = {
  /** The classname to be applied to the footer */
  className: PropTypes.string,
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
  /** Optional link for a support button */
  supportLink: PropTypes.string,
}

Footer.defaultProps = {
  className: undefined,
  links: [],
  supportLink: undefined,
}

export default Footer
