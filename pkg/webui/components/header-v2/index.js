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

import React, { useCallback } from 'react'
import classnames from 'classnames'
import { useDispatch, useSelector } from 'react-redux'

import { Breadcrumbs } from '@ttn-lw/components/breadcrumbs/breadcrumbs'
import Button from '@ttn-lw/components/button-v2'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown-v2'
import Dropdown from '@ttn-lw/components/dropdown'
import { BreadcrumbsConsumer, useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import selectAccountUrl from '@console/lib/selectors/app-config'
import { checkFromState, mayViewOrEditApiKeys } from '@console/lib/feature-checks'

import { logout } from '@account/store/actions/user'

import style from './header-v2.styl'

const accountUrl = selectAccountUrl()
const Header = ({ className, user, ...rest }) => {
  // Const isGuest = !Boolean(user)
  const dispatch = useDispatch()

  const mayHandleApiKeys = useSelector(state =>
    user ? checkFromState(mayViewOrEditApiKeys, state) : false,
  )

  const handleLogout = useCallback(() => {
    dispatch(logout())
  }, [dispatch])

  const dropdownItems = (
    <>
      <Dropdown.Item
        title={sharedMessages.profileSettings}
        icon="user"
        path={`${accountUrl}/profile-settings`}
        external
      />
      {mayHandleApiKeys && (
        <Dropdown.Item title={sharedMessages.apiKeys} icon="api_keys" path="/user/api-keys" />
      )}
      <Dropdown.Item
        title={sharedMessages.adminPanel}
        icon="lock"
        path="/admin-panel/network-information"
      />
      <hr />
      <Dropdown.Item
        title={sharedMessages.getSupport}
        icon="help"
        path="https://thethingsindustries.com/support"
        external
      />
      <Dropdown.Item
        title={sharedMessages.documentation}
        icon="description"
        path="https://thethingsindustries.com/docs"
        external
      />
      <hr />
      <Dropdown.Item title={sharedMessages.logout} icon="logout" action={handleLogout} />
    </>
  )

  return (
    <header {...rest} className={classnames(className, style.container)}>
      <BreadcrumbsConsumer>
        {({ breadcrumbs }) => <Breadcrumbs breadcrumbs={breadcrumbs} />}
      </BreadcrumbsConsumer>
      <div className={style.buttons}>
        <Button naked icon="add" withDropdown />
        <Button naked icon="grade" withDropdown />
        <Button naked icon="inbox" />
        <ProfileDropdown
          userName={user.name || user.ids.user_id}
          data-test-id="profile-dropdown"
          profilePicture={user.profile_picture}
        >
          {dropdownItems}
        </ProfileDropdown>
      </div>
    </header>
  )
}

Header.propTypes = {
  /** The classname applied to the component. */
  className: PropTypes.string,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered.
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  className: undefined,
  user: undefined,
}

export default Header
