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

import React from 'react'
import classnames from 'classnames'

import { IconStar, IconPlus, IconInbox, IconLayoutSidebarLeftExpand } from '@ttn-lw/components/icon'
import Button from '@ttn-lw/components/button'
import ProfileDropdown from '@ttn-lw/components/profile-dropdown'

import AppStatusBadge from '@console/containers/app-status-badge'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import Link from '../link'

import style from './header.styl'

const Header = ({
  brandLogo,
  Logo,
  safe,
  className,
  addDropdownItems,
  bookmarkDropdownItems,
  profileDropdownItems,
  notificationsDropdownItems,
  user,
  onMenuClick,
  showNotificationDot,
  alwaysShowLogo,
  isSidebarMinimized,
  toggleSidebarMinimized,
  expandSidebar,
  handleHideSidebar,
  hasCreateRights,
  ...rest
}) => {
  const LinkComponent = safe ? 'a' : Link
  const darkTheme = window.matchMedia('(prefers-color-scheme: dark)').matches

  return (
    <header
      {...rest}
      className={classnames(className, style.container, {
        [style.containerMinimized]: isSidebarMinimized,
      })}
      id="header"
    >
      {alwaysShowLogo ? (
        <div className="d-flex al-center gap-cs-xs">
          <LinkComponent to="/" href="/">
            <Logo className={style.logo} dark={darkTheme} />
          </LinkComponent>
        </div>
      ) : (
        <>
          <div className="d-flex j-start al-center gap-cs-s lg-xl:d-none">
            {isSidebarMinimized && (
              <>
                <Button
                  className="md-lg:d-none"
                  icon={IconLayoutSidebarLeftExpand}
                  onClick={toggleSidebarMinimized}
                  naked
                  tooltip={sharedMessages.keepSidebarOpen}
                  tooltipPlacement="right"
                  onMouseEnter={expandSidebar}
                  onMouseLeave={handleHideSidebar}
                />
                <div className={style.divider} />
              </>
            )}
            <div className={classnames('breadcrumbs', 'lg-xl:d-none')} />
          </div>
          <div className="d-none lg-xl:d-flex al-center gap-cs-xs">
            <Button secondary icon={IconLayoutSidebarLeftExpand} onClick={onMenuClick} />
            <LinkComponent to="/" href="/">
              <Logo className={style.logo} dark={darkTheme} />
            </LinkComponent>
          </div>
        </>
      )}

      {!safe && (
        <div className="d-flex al-center gap-cs-xs">
          <AppStatusBadge />
          {hasCreateRights && (
            <Button
              primary
              message={sharedMessages.add}
              icon={IconPlus}
              dropdownItems={addDropdownItems}
              dropdownPosition="below left"
              tooltip={sharedMessages.addEntity}
              tooltipPlacement="bottom"
              className="md-lg:d-none"
            />
          )}

          <Button
            secondary
            icon={IconStar}
            dropdownItems={bookmarkDropdownItems}
            dropdownClassName={style.bookmarksDropdown}
            dropdownPosition="below left"
            tooltip={sharedMessages.bookmarks}
            tooltipPlacement="bottom"
            className="md-lg:d-none"
          />
          <Button
            secondary
            icon={IconInbox}
            dropdownItems={notificationsDropdownItems}
            dropdownClassName={style.notificationsDropdown}
            dropdownPosition="below left"
            tooltip={sharedMessages.notifications}
            tooltipPlacement="bottom"
            withAlert={showNotificationDot}
            className="md-lg:d-none"
          />
          <ProfileDropdown
            brandLogo={brandLogo}
            data-test-id="profile-dropdown"
            profilePicture={user?.profile_picture}
          >
            {profileDropdownItems}
          </ProfileDropdown>
        </div>
      )}
    </header>
  )
}

const imgPropType = PropTypes.shape({
  src: PropTypes.string.isRequired,
  alt: PropTypes.string.isRequired,
})

Header.propTypes = {
  /** The logo component. */
  Logo: PropTypes.elementType.isRequired,
  /** The dropdown items when the add button is clicked. */
  addDropdownItems: PropTypes.node,
  /** Whether to always show the logo, which is required in error views, where there is no sidebar. */
  alwaysShowLogo: PropTypes.bool,
  /** The dropdown items when the bookmark button is clicked. */
  bookmarkDropdownItems: PropTypes.node,
  brandLogo: imgPropType,
  /** The classname applied to the component. */
  className: PropTypes.string,
  /** A handler for when the sidebar is expanded. */
  expandSidebar: PropTypes.func.isRequired,
  /** A handler for when the sidebar is hidden. */
  handleHideSidebar: PropTypes.func,
  /** Has the right to create entities. */
  hasCreateRights: PropTypes.bool,
  /** Whether the sidebar is minimized. */
  isSidebarMinimized: PropTypes.bool.isRequired,
  /** The dropdown items when the notifications button is clicked. */
  notificationsDropdownItems: PropTypes.node,
  /** A handler for when the menu button is clicked. */
  onMenuClick: PropTypes.func,
  /** The dropdown items when the profile button is clicked. */
  profileDropdownItems: PropTypes.node,
  /** Whether the header should be rendered in safe mode. */
  safe: PropTypes.bool,
  /** Whether to show a notification dot. */
  showNotificationDot: PropTypes.bool,
  toggleSidebarMinimized: PropTypes.func.isRequired,
  /**
   * The User object, retrieved from the API. If it is `undefined`, then the
   * guest header is rendered.
   */
  user: PropTypes.user,
}

Header.defaultProps = {
  alwaysShowLogo: false,
  className: undefined,
  user: undefined,
  brandLogo: undefined,
  showNotificationDot: false,
  safe: false,
  addDropdownItems: undefined,
  bookmarkDropdownItems: undefined,
  notificationsDropdownItems: undefined,
  profileDropdownItems: undefined,
  onMenuClick: () => null,
  handleHideSidebar: () => null,
  hasCreateRights: false,
}

export default Header
