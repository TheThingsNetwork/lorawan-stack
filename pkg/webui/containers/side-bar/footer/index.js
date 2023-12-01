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

import React, { useState } from 'react'
import classNames from 'classnames'

import Button from '@ttn-lw/components/button-v2'
import Dropdown from '@ttn-lw/components/dropdown-v2'

import {
  selectDocumentationUrlConfig,
  selectPageStatusBaseUrlConfig,
  selectSupportLinkConfig,
} from '@ttn-lw/lib/selectors/env'

import style from './side-footer.styl'

const supportLink = selectSupportLinkConfig()
const documentationBaseUrl = selectDocumentationUrlConfig()
const statusPageBaseUrl = selectPageStatusBaseUrlConfig()

const clusterDropdownItems = (
  <>
    <Dropdown.Item title="Cluster" icon="public" />
  </>
)

const supportDropdownItems = (
  <>
    <Dropdown.Item title="Cluster" icon="public" />
  </>
)

const SideFooter = () => {
  const [isExpanded, setIsExpanded] = useState(false)

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
        isHoverDropdown
      />
      <Button
        secondary
        withDropdown
        icon="public"
        message="EU1"
        dropdownItems={clusterDropdownItems}
      />
    </div>
  )
}

export default SideFooter
