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
import classNames from 'classnames'

import Button from '@ttn-lw/components/button-v2'
import Dropdown from '@ttn-lw/components/dropdown-v2'

import SideBarContext from '@ttn-lw/containers/side-bar/context'

import { LanguageContext } from '@ttn-lw/lib/components/with-locale'

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
  const { isMinimized } = useContext(SideBarContext)
  const ref = useRef(null)

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
      <Dropdown.Item title="Documentation" icon="menu_book" path={documentationBaseUrl} />
      <Dropdown.Item title="Support" icon="support" path={supportLink} />
      <Dropdown.Item title="Status page" icon="monitor_heart" path={statusPageBaseUrl} />
      {Boolean(languageContext) && (
        <Dropdown.Item
          title="Language"
          icon="language"
          path="/support"
          submenuItems={submenuItems}
        />
      )}
    </>
  )

  return (
    <div
      className={classNames(
        style.sideFooter,
        'd-flex',
        'j-between',
        'align-center',
        'gap-cs-xs',
        'fs-xs',
        'w-90',
      )}
    >
      <Button
        className={style.sideFooterButton}
        secondary
        message={!isMinimized ? `v${process.env.VERSION} (${process.env.REVISION})` : undefined}
        icon="support"
        dropdownItems={supportDropdownItems}
        dropdownClassName={style.sideFooterHoverDropdown}
        isHoverDropdown
        ref={ref}
      />
      {!isMinimized && (
        <Button
          secondary
          withDropdown
          icon="public"
          message="EU1"
          dropdownItems={clusterDropdownItems}
          dropdownClassName={style.sideFooterDropdown}
          ref={ref}
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
