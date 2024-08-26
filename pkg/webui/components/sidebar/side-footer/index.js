// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { defineMessages } from 'react-intl'

import {
  IconHeartRateMonitor,
  IconLanguage,
  IconBook,
  IconSupport,
  IconInfoSquareRounded,
} from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Dropdown from '@ttn-lw/components/dropdown'

import { LanguageContext } from '@ttn-lw/lib/components/with-locale'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import {
  selectDocumentationUrlConfig,
  selectPageStatusBaseUrlConfig,
  selectSupportLinkConfig,
} from '@ttn-lw/lib/selectors/env'

import style from './side-footer.styl'

const m = defineMessages({
  resources: 'Resources',
  clusterSelection: 'Cluster selection',
})

const supportLink = selectSupportLinkConfig()
const documentationBaseUrl = selectDocumentationUrlConfig()
const statusPageBaseUrl = selectPageStatusBaseUrlConfig()

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

const SideFooter = () => {
  const supportButtonRef = useRef(null)

  const languageContext = useContext(LanguageContext)
  const { locale, supportedLocales, setLocale } = languageContext || {}

  const handleSetLocale = useCallback(
    locale => {
      setLocale(locale)
    },
    [setLocale],
  )

  const languageItems = supportedLocales
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
        icon={IconBook}
        path={documentationBaseUrl}
        external
      />
      <Dropdown.Item
        title={sharedMessages.support}
        icon={IconSupport}
        path={supportLink}
        external
      />
      <Dropdown.Item
        title={sharedMessages.statusPage}
        icon={IconHeartRateMonitor}
        path={statusPageBaseUrl}
        external
      />
      {Boolean(languageContext) && (
        <Dropdown.Item
          title={sharedMessages.language}
          icon={IconLanguage}
          path="/support"
          submenuItems={languageItems}
          external
        />
      )}
    </>
  )

  const sideFooterClassnames = classnames('d-flex', 'j-between', 'al-center', 'gap-cs-m', 'fs-s')

  return (
    <div className={style.sideFooter}>
      <div className={sideFooterClassnames}>
        <Button
          className={style.supportButton}
          secondary
          message={m.resources}
          icon={IconInfoSquareRounded}
          dropdownItems={supportDropdownItems}
          dropdownPosition="above"
          dropdownClassName={style.sideFooterDropdown}
          ref={supportButtonRef}
        />
        <Button className={style.clusterButton} noDropdownIcon>
          <span className={style.clusterButtonContent}>
            <span className={style.sideFooterVersion}>
              v{process.env.VERSION}.{process.env.REVISION}
            </span>
          </span>
        </Button>
      </div>
    </div>
  )
}

export default SideFooter
