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

import React, { useCallback, useContext, useRef } from 'react'
import classnames from 'classnames'

import Button from '@ttn-lw/components/button-v2'
import Dropdown from '@ttn-lw/components/dropdown-v2'

import SidebarContext from '@ttn-lw/containers/side-bar/context'

import { LanguageContext } from '@ttn-lw/lib/components/with-locale'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './side-footer.styl'

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

const SideFooter = ({ supportLink, documentationBaseUrl, statusPageBaseUrl }) => {
  const { isMinimized } = useContext(SidebarContext)
  const supportButtonRef = useRef(null)
  const clusterButtonRef = useRef(null)

  const clusterDropdownItems = (
    <>
      <Dropdown.Item title="Cluster selection" icon="public" path="/cluster" />
    </>
  )

  const languageContext = useContext(LanguageContext)
  const { locale, supportedLocales, setLocale } = languageContext || {}

  const handleSetLocale = useCallback(
    locale => {
      setLocale(locale)
    },
    [setLocale],
  )

  const submenuItems = supportedLocales
    ? Object.keys(supportedLocales).map(l => (
        <LanguageOption
          locale={l}
          key={l}
          title={supportedLocales[l]}
          currentLocale={locale}
          onSetLocale={handleSetLocale}
        />
      ))
    : null

  const supportDropdownItems = (
    <>
      <Dropdown.Item
        title={sharedMessages.documentation}
        icon="menu_book"
        path={documentationBaseUrl}
      />
      <Dropdown.Item title={sharedMessages.support} icon="support" path={supportLink} />
      <Dropdown.Item
        title={sharedMessages.statusPage}
        icon="monitor_heart"
        path={statusPageBaseUrl}
      />
      {Boolean(languageContext) && (
        <Dropdown.Item
          title={sharedMessages.language}
          icon="language"
          path="/support"
          submenuItems={submenuItems}
        />
      )}
    </>
  )

  const sideFooterClassnames = classnames(
    style.sideFooter,
    'd-flex',
    'j-center',
    'al-center',
    'gap-cs-xs',
    'fs-s',
  )

  return (
    <div className={sideFooterClassnames}>
      <Button
        className={classnames(style.sideFooterButton, style.sideFooterFullWidth)}
        secondary
        icon="support"
        message={!isMinimized ? `v${process.env.VERSION} (${process.env.REVISION})` : undefined}
        dropdownItems={supportDropdownItems}
        dropdownClassName={classnames(style.sideFooterDropdown, style.sideFooterSupportDropdown)}
        noDropdownIcon
        ref={supportButtonRef}
      />
      {!isMinimized && (
        <Button
          className={style.sideFooterButton}
          secondary
          icon="public"
          message="EU1"
          noDropdownIcon
          dropdownItems={clusterDropdownItems}
          dropdownClassName={style.sideFooterDropdown}
          ref={clusterButtonRef}
        />
      )}
    </div>
  )
}

SideFooter.propTypes = {
  documentationBaseUrl: PropTypes.string.isRequired,
  statusPageBaseUrl: PropTypes.string.isRequired,
  supportLink: PropTypes.string.isRequired,
}

export default SideFooter
