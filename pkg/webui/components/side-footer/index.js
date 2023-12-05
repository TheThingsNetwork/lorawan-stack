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

import React, { useRef } from 'react'
import classNames from 'classnames'

import Button from '@ttn-lw/components/button-v2'
import Dropdown from '@ttn-lw/components/dropdown-v2'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './side-footer.styl'

const SideFooter = ({ supportLink, documentationBaseUrl, statusPageBaseUrl }) => {
  const ref = useRef(null)

  const clusterDropdownItems = (
    <>
      <Dropdown.Item title="Cluster selection" icon="public" path="/cluster" />
    </>
  )

  const submenuItems = (
    <>
      <Dropdown.Item title="EN" />
      <Dropdown.Item title="JP" />
    </>
  )

  const supportDropdownItems = (
    <>
      <Dropdown.Item title="Documentation" icon="menu_book" path={documentationBaseUrl} />
      <Dropdown.Item title="Support" icon="support" path={supportLink} />
      <Dropdown.Item title="Status page" icon="monitor_heart" path={statusPageBaseUrl} />
      <Dropdown.Item title="Language" icon="language" path="/support" submenuItems={submenuItems} />
    </>
  )

  return (
    <div
      className={classNames(
        style.sideFooter,
        'd-flex',
        'j-center',
        'align-center',
        'gap-cs-xxs',
        'fs-xs',
        'w-90',
      )}
    >
      <Button
        className={style.sideFooterButton}
        secondary
        message={`v${process.env.VERSION} (${process.env.REVISION})`}
        icon="support"
        dropdownItems={supportDropdownItems}
        dropdownClassName={style.sideFooterHoverDropdown}
        isHoverDropdown
        ref={ref}
      />
      <Button
        secondary
        withDropdown
        icon="public"
        message="EU1"
        dropdownItems={clusterDropdownItems}
        dropdownClassName={style.sideFooterDropdown}
        ref={ref}
      />
    </div>
  )
}

SideFooter.propTypes = {
  documentationBaseUrl: PropTypes.string.isRequired,
  statusPageBaseUrl: PropTypes.string.isRequired,
  supportLink: PropTypes.string.isRequired,
}

export default SideFooter
