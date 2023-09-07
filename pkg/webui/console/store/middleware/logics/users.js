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

import { defineMessages } from 'react-intl'

import tts from '@console/api/tts'

import toast from '@ttn-lw/components/toast'

import createRequestLogic from '@ttn-lw/lib/store/logics/create-request-logic'
import { getBackendErrorName } from '@ttn-lw/lib/errors/utils'

import * as users from '@console/store/actions/users'

const getUserLogic = createRequestLogic({
  type: users.GET_USER,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { selector },
    } = action

    return await tts.Users.getById(id, selector)
  },
})

const updateUserLogic = createRequestLogic({
  type: users.UPDATE_USER,
  process: ({ action }) => {
    const {
      payload: { id, patch },
    } = action

    return tts.Users.updateById(id, patch)
  },
})

const deleteUserLogic = createRequestLogic({
  type: users.DELETE_USER,
  process: async ({ action }) => {
    const {
      payload: { id },
      meta: { options },
    } = action

    if (options.purge) {
      await tts.Users.purgeById(id)
    } else {
      await tts.Users.deleteById(id)
    }

    return { id }
  },
})

const restoreUserLogic = createRequestLogic({
  type: users.RESTORE_USER,
  process: async ({ action }) => {
    const { id } = action.payload

    await tts.Users.restoreById(id)

    return { id }
  },
})

const getUsersLogic = createRequestLogic({
  type: users.GET_USERS_LIST,
  process: async ({ action }) => {
    const {
      params: { page, limit, query, order, deleted },
    } = action.payload
    const { selectors, options } = action.meta

    const data = options.isSearch
      ? await tts.Users.search(
          {
            page,
            limit,
            query,
            order,
            deleted,
          },
          selectors,
        )
      : await tts.Users.getAll({ page, limit, order }, selectors)

    return { entities: data.users, totalCount: data.totalCount }
  },
})

const createUserLogic = createRequestLogic({
  type: users.CREATE_USER,
  process: async ({ action }) => {
    const {
      payload: { user },
    } = action

    return await tts.Users.create(user)
  },
})

const getUsersRightsLogic = createRequestLogic({
  type: users.GET_USER_RIGHTS_LIST,
  process: async ({ action }) => {
    const { id } = action.payload
    const result = await tts.Users.getRightsById(id)

    return result.rights.sort()
  },
})

const getUserInvitationsLogic = createRequestLogic({
  type: users.GET_USER_INVITATIONS,
  process: async ({ action }) => {
    const {
      params: { page, limit },
    } = action.payload
    const { selectors } = action.meta

    return await tts.Users.getAllInvitations({ page, limit }, selectors)
  },
})

const sendInviteLogic = createRequestLogic({
  type: users.SEND_INVITE,
  process: async ({ action }) => {
    const { email } = action.payload

    return await tts.Users.sendInvite(email)
  },
})

const deleteInviteLogic = createRequestLogic({
  type: users.DELETE_INVITE,
  process: async ({ action }) => {
    const { email } = action.payload

    return await tts.Users.deleteInvite(email)
  },
})

const m = defineMessages({
  errEmailValidationActionSuccess: 'Validation email sent (please also check your spam folder)',
  errEmailValidationActionFailure: 'There was an error and the validation email could not be sent.',
  errEmailValidationAlreadySent:
    'A validation email has already been sent recently to your email address. Please also check your spam folder.',
})

const requestEmailValidationLogic = createRequestLogic({
  type: users.REQUEST_EMAIL_VALIDATION,
  process: async ({ action }) => {
    const { userId } = action.payload
    try {
      const result = await tts.ContactInfo.requestValidation({ user_ids: { user_id: userId } })
      toast({
        type: toast.types.SUCCESS,
        message: m.errEmailValidationActionSuccess,
      })
      return result
    } catch (error) {
      if (getBackendErrorName(error) === 'validations_already_sent') {
        toast({
          type: toast.types.INFO,
          message: m.errEmailValidationAlreadySent,
        })
        return
      }
      toast({
        type: toast.types.ERROR,
        message: m.errEmailValidationActionFailure,
      })
    }
  },
})

export default [
  getUserLogic,
  getUsersLogic,
  updateUserLogic,
  deleteUserLogic,
  restoreUserLogic,
  createUserLogic,
  getUsersRightsLogic,
  getUserInvitationsLogic,
  sendInviteLogic,
  deleteInviteLogic,
  requestEmailValidationLogic,
]
