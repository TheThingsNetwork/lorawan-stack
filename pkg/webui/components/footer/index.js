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

import React, { useContext, useCallback, useState, useRef } from 'react'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'
import Link from '@ttn-lw/components/link'
import OfflineStatus from '@ttn-lw/components/offline-status'
import Dropdown from '@ttn-lw/components/dropdown'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import { LanguageContext } from '@ttn-lw/lib/components/with-locale'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './footer.styl'

const year = new Date().getFullYear()

const LanguageOption = ({ locale, title, currentLocale, onSetLocale }) => {
  const handleSetLocale = useCallback(() => {
    onSetLocale(locale)
  }, [locale, onSetLocale])

  return <Dropdown.Item title={title} action={handleSetLocale} active={locale === currentLocale} />
}

LanguageOption.propTypes = {
  currentLocale: PropTypes.string.isRequired,
  locale: PropTypes.string.isRequired,
  onSetLocale: PropTypes.func.isRequired,
  title: PropTypes.string.isRequired,
}

const FooterSection = ({ className, action, link, children, primary, ...rest }) => {
  let content
  if (Boolean(link)) {
    content = (
      <Button.AnchorLink
        className={style.footerSectionButton}
        href={link}
        secondary
        unstyled
        target="blank"
      >
        {children}
      </Button.AnchorLink>
    )
  } else if (Boolean(action)) {
    content = (
      <Button className={style.footerSectionButton} onClick={action} unstyled responsiveLabel>
        {children}
      </Button>
    )
  } else {
    content = children
  }

  const cls = classnames(className, style.footerSection, {
    [style.interactive]: Boolean(action || link),
    [style.primary]: primary,
  })

  return <div className={cls}>{content}</div>
}

FooterSection.propTypes = {
  action: PropTypes.func,
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  link: PropTypes.string,
  primary: PropTypes.bool,
}

FooterSection.defaultProps = {
  action: undefined,
  className: undefined,
  link: undefined,
  primary: false,
}

const Footer = ({
  className,
  documentationLink,
  links,
  supportLink,
  onlineStatus,
  transparent,
}) => {
  const { locale, supportedLocales, setLocale } = useContext(LanguageContext)
  const [languageDropdownVisible, setLanguageDropdownVisible] = useState(false)
  const node = useRef(null)

  const handleClickOutside = useCallback(
    e => {
      if (!node.current || !node.current.contains(e.target)) {
        setLanguageDropdownVisible(false)
      }
    },
    [setLanguageDropdownVisible],
  )

  const showLanguageDropdown = useCallback(() => {
    document.addEventListener('mousedown', handleClickOutside)
    setLanguageDropdownVisible(true)
  }, [setLanguageDropdownVisible, handleClickOutside])

  const hideLanguageDropdown = useCallback(() => {
    document.removeEventListener('mousedown', handleClickOutside)
    setLanguageDropdownVisible(false)
  }, [setLanguageDropdownVisible, handleClickOutside])

  const handleToggleLanguageDropdown = useCallback(() => {
    if (languageDropdownVisible) {
      hideLanguageDropdown()
    } else {
      showLanguageDropdown()
    }
  }, [hideLanguageDropdown, showLanguageDropdown, languageDropdownVisible])

  const handleSetLocale = useCallback(
    locale => {
      setLocale(locale)
      hideLanguageDropdown()
    },
    [setLocale, hideLanguageDropdown],
  )

  return (
    <footer className={classnames(className, style.footer, { [style.transparent]: transparent })}>
      <div className={style.left}>
        <div>
          © {year}{' '}
          <Link.Anchor
            secondary
            className={style.link}
            href="https://www.thethingsindustries.com/docs"
          >
            The Things Stack
          </Link.Anchor>{' '}
          <span className={style.copyrightLinks}>
            by{' '}
            <Link.Anchor secondary className={style.link} href="https://www.thethingsnetwork.org">
              The Things Network
            </Link.Anchor>{' '}
            and{' '}
            <Link.Anchor
              secondary
              className={style.link}
              href="https://www.thethingsindustries.com"
            >
              The Things Industries
            </Link.Anchor>
          </span>
        </div>
      </div>
      <div className={style.right}>
        {links.map((item, key) => (
          <Link.Anchor secondary key={key} className={style.link} href={item.link}>
            <Message content={item.title} />
          </Link.Anchor>
        ))}
        <OfflineStatus onlineStatus={onlineStatus} showOfflineOnly showWarnings />
        <div className={style.language} ref={node}>
          <FooterSection action={handleToggleLanguageDropdown} icon="language">
            <Icon icon="language" className={style.languageIcon} textPaddedRight />
            {locale.split('-')[0].toUpperCase()}
          </FooterSection>
          {languageDropdownVisible && (
            <Dropdown className={style.languageDropdown}>
              {Object.keys(supportedLocales).map(l => (
                <LanguageOption
                  locale={l}
                  key={l}
                  title={supportedLocales[l]}
                  currentLocale={locale}
                  onSetLocale={handleSetLocale}
                />
              ))}
            </Dropdown>
          )}
        </div>
        <FooterSection link={documentationLink ? `${documentationLink}/whats-new/` : undefined}>
          v{process.env.VERSION}
        </FooterSection>
        {documentationLink && (
          <FooterSection className={style.documentation} link={documentationLink}>
            <Message content={sharedMessages.documentation} />
          </FooterSection>
        )}
        {true && (
          <FooterSection link={'http://localhost'} primary>
            <Icon icon="contact_support" textPaddedRight nudgeDown />
            <Message content={sharedMessages.getSupport} />
          </FooterSection>
        )}
      </div>
    </footer>
  )
}

Footer.propTypes = {
  /** The classname to be applied to the footer. */
  className: PropTypes.string,
  /** Optional link for documentation docs. */
  documentationLink: PropTypes.string,
  /**
   * A list of links to be displayed in the footer component.
   *
   * @param {(string|object)} title - The title of the link.
   * @param {string} link - The link url.
   */
  links: PropTypes.arrayOf(
    PropTypes.shape({
      title: PropTypes.message.isRequired,
      link: PropTypes.string.isRequired,
    }),
  ),
  /** A flag specifying whether the application is connected to the internet. */
  onlineStatus: PropTypes.onlineStatus.isRequired,
  /** Optional link for a support button. */
  supportLink: PropTypes.string,
  /** Whether transparent styling should be applied. */
  transparent: PropTypes.bool,
}

Footer.defaultProps = {
  className: undefined,
  documentationLink: undefined,
  links: [],
  supportLink: undefined,
  transparent: false,
}

export default Footer
