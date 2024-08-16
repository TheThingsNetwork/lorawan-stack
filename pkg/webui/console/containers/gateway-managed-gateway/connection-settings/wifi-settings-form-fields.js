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

import React, { useCallback, useMemo } from 'react'
import { defineMessages } from 'react-intl'
import { useDispatch, useSelector } from 'react-redux'
import { isEmpty } from 'lodash'
import { useParams } from 'react-router-dom'
import classNames from 'classnames'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Select from '@ttn-lw/components/select'
import Icon from '@ttn-lw/components/icon'
import Notification from '@ttn-lw/components/notification'
import Button from '@ttn-lw/components/button'
import Link from '@ttn-lw/components/link'
import toast from '@ttn-lw/components/toast'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import GatewayWifiProfilesFormFields from '@console/containers/gateway-managed-gateway/shared/wifi-profiles-form-fields'
import ShowProfilesSelect from '@console/containers/gateway-managed-gateway/shared/show-profiles-select'
import {
  CONNECTION_TYPES,
  initialWifiProfile,
} from '@console/containers/gateway-managed-gateway/shared/utils'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import PropTypes from '@ttn-lw/lib/prop-types'
import diff from '@ttn-lw/lib/diff'

import { getConnectionProfilesList } from '@console/store/actions/connection-profiles'

import { selectConnectionProfilesByType } from '@console/store/selectors/connection-profiles'

import style from './wifi-settings-form-fields.styl'

const m = defineMessages({
  settingsProfile: 'Settings profile',
  profileDescription: 'Connection settings profiles can be shared within the same organization',
  wifiConnection: 'WiFi connection',
  selectAProfile: 'Select a profile',
  connected: 'The gateway WiFi successfully connected using this profile',
  connectedCollaborator: "The gateway WiFi successfully connected using a collaborator's profile",
  unableToConnect: 'The gateway WiFi is currently unable to connect using this profile',
  unableToConnectCollaborator:
    "The gateway WiFi is currently unable to connect using  a collaborator's profile",
  saveToConnect: 'Please click "Save changes" to start using this WiFi profile for the gateway',
  createNewSharedProfile: 'Create a new shared profile',
  setAConfigForThisGateway: 'Set a config for this gateway only',
  notificationContent:
    'This gateway already has a WiFi profile set by another collaborator. If wished, you can override this profile below.',
  overrideProfile: 'Override this profile',
  editProfile: 'Edit this profile',
  attemptingToConnect: 'The gateway WiFi is currently attempting to connect using this profile',
  settingsProfileTooltip:
    'To set up the gateway connection, you can either use a shared profile, to share the connection settings with other gateways, or set a config for this gateway only.',
  fetchProfilesFailure: 'There was an error and the WiFi profiles cannot be fetched.',
})

const WifiSettingsFormFields = ({ initialValues, isWifiConnected, saveFormClicked }) => {
  const { gtwId } = useParams()
  const { values, setValues } = useFormContext()
  const dispatch = useDispatch()
  const profiles = useSelector(state =>
    selectConnectionProfilesByType(state, CONNECTION_TYPES.WIFI),
  )

  const hasChanged = useMemo(
    () =>
      !isEmpty(
        diff(initialValues.wifi_profile, values.wifi_profile, {
          exclude: ['_profile_of', '_access_point'],
        }),
      ),
    [initialValues, values],
  )

  const profileOptions = [
    ...profiles.map(p => ({
      value: p.profile_id,
      label: p.profile_name,
    })),
    { value: 'shared', label: m.createNewSharedProfile.defaultMessage },
    { value: 'non-shared', label: m.setAConfigForThisGateway.defaultMessage },
  ]

  const connectionStatus = useMemo(() => {
    if (!values.wifi_profile.profile_id) return null
    if (hasChanged) {
      return { message: m.saveToConnect, icon: 'info_outline' }
    }
    if (isWifiConnected) {
      if (values.wifi_profile._override) {
        return {
          message: m.connectedCollaborator,
          icon: 'check_circle_outline',
          color: style.connected,
        }
      }
      return { message: m.connected, icon: 'check_circle_outline', color: style.connected }
    }
    if (!isWifiConnected) {
      if (saveFormClicked) {
        return {
          message: m.attemptingToConnect,
          icon: 'rotate_right',
        }
      }
      if (values.wifi_profile._override) {
        return {
          message: m.unableToConnectCollaborator,
          icon: 'highlight_remove',
          color: style.notConnected,
        }
      }
      return { message: m.unableToConnect, icon: 'highlight_remove', color: style.notConnected }
    }

    return null
  }, [
    hasChanged,
    isWifiConnected,
    saveFormClicked,
    values.wifi_profile._override,
    values.wifi_profile.profile_id,
  ])

  const handleChangeProfile = useCallback(
    async value => {
      setValues(values => ({
        ...values,
        wifi_profile: {
          ...values.wifi_profile,
          profile_id: '',
        },
      }))
      try {
        await dispatch(
          attachPromise(
            getConnectionProfilesList({
              entityId: value,
              type: CONNECTION_TYPES.WIFI,
            }),
          ),
        )
      } catch (e) {
        toast({
          message: m.fetchProfilesFailure,
          type: toast.types.ERROR,
        })
      }
    },
    [dispatch, setValues],
  )

  const handleOverrideProfile = useCallback(
    () =>
      setValues(oldValues => ({
        ...oldValues,
        wifi_profile: {
          ...oldValues.wifi_profile,
          _override: false,
          profile_id: '',
        },
      })),
    [setValues],
  )

  const handleProfileIdChange = useCallback(
    value => {
      if (value.includes('shared')) {
        const { profile_id, _profile_of, ...initialProfile } = initialWifiProfile
        setValues(oldValues => ({
          ...oldValues,
          wifi_profile: {
            ...oldValues.wifi_profile,
            ...initialProfile,
            profile_name: value === 'non-shared' ? Date.now().toString() : '',
          },
        }))
      }
    },
    [setValues],
  )

  return (
    <>
      <Message component="h3" content={m.wifiConnection} className="mt-0" />
      {!values.wifi_profile._override ? (
        <>
          <div className="d-flex al-center gap-cs-m">
            <ShowProfilesSelect name={`wifi_profile._profile_of`} onChange={handleChangeProfile} />
            {Boolean(values.wifi_profile._profile_of) && (
              <RequireRequest
                requestAction={getConnectionProfilesList({
                  entityId: values.wifi_profile._profile_of,
                  type: CONNECTION_TYPES.WIFI,
                })}
                handleErrors={false}
              >
                <Form.Field
                  name={`wifi_profile.profile_id`}
                  title={m.settingsProfile}
                  component={Select}
                  options={profileOptions}
                  tooltip={m.settingsProfileTooltip}
                  placeholder={m.selectAProfile}
                  onChange={handleProfileIdChange}
                />
              </RequireRequest>
            )}
          </div>
          <Message
            component="div"
            content={m.profileDescription}
            className={style.fieldDescription}
          />
          {values.wifi_profile.profile_id.includes('shared') && (
            <GatewayWifiProfilesFormFields namePrefix={`wifi_profile.`} />
          )}
        </>
      ) : (
        <div>
          <Notification info small content={m.notificationContent} />
          <Button
            type="button"
            className="mb-cs-m"
            message={m.overrideProfile}
            onClick={handleOverrideProfile}
          />
        </div>
      )}

      {connectionStatus !== null && (
        <div className="d-inline-flex al-center gap-cs-m">
          <div className={classNames(style.connection, connectionStatus.color)}>
            <Icon icon={connectionStatus.icon} className={connectionStatus.color} />
            <Message content={connectionStatus.message} />
          </div>
          {Boolean(values.wifi_profile.profile_id) &&
            !values.wifi_profile._override &&
            !values.wifi_profile.profile_id.includes('shared') && (
              <Link
                primary
                to={`/gateways/${gtwId}/managed-gateway/wifi-profiles/edit/${values.wifi_profile.profile_id}?profileOf=${values.wifi_profile._profile_of}`}
              >
                <Message content={m.editProfile} />
              </Link>
            )}
        </div>
      )}
    </>
  )
}

WifiSettingsFormFields.propTypes = {
  initialValues: PropTypes.shape({
    wifi_profile: PropTypes.shape({
      _override: PropTypes.bool,
      profile_id: PropTypes.string,
      _profile_of: PropTypes.string,
    }),
  }).isRequired,
  isWifiConnected: PropTypes.bool,
  saveFormClicked: PropTypes.bool.isRequired,
}

WifiSettingsFormFields.defaultProps = {
  isWifiConnected: false,
}

export default WifiSettingsFormFields
