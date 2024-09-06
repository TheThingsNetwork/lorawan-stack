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

import React from 'react'
import classnames from 'classnames'
import { useDispatch, useSelector } from 'react-redux'

import { APPLICATION, END_DEVICE } from '@console/constants/entities'

import { IconApplication, IconDevice, IconGateway, IconOrganization } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import Dropdown from '@ttn-lw/components/dropdown'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayViewApplications,
  mayViewGateways,
  mayViewOrganizationsOfUser,
} from '@console/lib/feature-checks'

import { setSearchOpen, setSearchScope } from '@console/store/actions/search'

import { selectUser } from '@console/store/selectors/user'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './section-label.styl'

const SectionLabel = ({
  label,
  icon,
  className,
  buttonDisabled,
  type,
  'data-test-id': dataTestId,
}) => {
  const dispatch = useDispatch()
  const user = useSelector(selectUser)
  const appId = useSelector(selectSelectedApplicationId)
  const mayViewApps = useSelector(state =>
    user ? checkFromState(mayViewApplications, state) : false,
  )
  const mayViewGtws = useSelector(state => (user ? checkFromState(mayViewGateways, state) : false))
  const mayViewOrgs = useSelector(state =>
    user ? checkFromState(mayViewOrganizationsOfUser, state) : false,
  )

  const handleRegisterDeviceClick = React.useCallback(() => {
    dispatch(setSearchScope(APPLICATION))
    dispatch(setSearchOpen(true))
  }, [dispatch])

  const plusDropdownItems = (
    <>
      {mayViewApps && (
        <Dropdown.Item
          title={sharedMessages.addApplication}
          icon={IconApplication}
          path="/applications/add"
        />
      )}
      {mayViewGtws && (
        <Dropdown.Item title={sharedMessages.addGateway} icon={IconGateway} path="/gateways/add" />
      )}
      {mayViewOrgs && (
        <Dropdown.Item
          title={sharedMessages.addOrganization}
          icon={IconOrganization}
          path="/organizations/add"
        />
      )}

      <Dropdown.Item
        title={
          type === END_DEVICE
            ? sharedMessages.registerEndDevice
            : sharedMessages.registerDeviceInApplication
        }
        icon={IconDevice}
        path={type === END_DEVICE ? `/applications/${appId}/devices/add` : undefined}
        action={type === END_DEVICE ? undefined : handleRegisterDeviceClick}
      />
    </>
  )

  return (
    <div
      className={classnames(
        className,
        'd-flex',
        'j-between',
        'al-center',
        'c-text-neutral-light',
        'ml-cs-xs',
        'fs-s',
      )}
      data-test-id={dataTestId}
    >
      <Message content={label} />
      <Button
        naked
        small
        icon={icon}
        disabled={buttonDisabled}
        dropdownItems={plusDropdownItems}
        dropdownPosition="below"
        noDropdownIcon
        dropdownClassName={style.portalledDropdown}
        portalledDropdown
        tooltip={sharedMessages.addEntity}
        tooltipPlacement="bottom"
      />
    </div>
  )
}

SectionLabel.propTypes = {
  buttonDisabled: PropTypes.bool,
  className: PropTypes.string,
  'data-test-id': PropTypes.string,
  icon: PropTypes.icon.isRequired,
  label: PropTypes.oneOfType([PropTypes.node, PropTypes.message]).isRequired,
  type: PropTypes.string,
}

SectionLabel.defaultProps = {
  buttonDisabled: false,
  className: undefined,
  'data-test-id': 'section-label',
  type: undefined,
}

export default SectionLabel
