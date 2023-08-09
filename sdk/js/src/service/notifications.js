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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

class Notifications {
  constructor(service) {
    this._api = service
    autoBind(this)
  }

  async getAllNotifications(recieverId, selector) {
    const result = await this._api.NotificationService.List(
      {
        routeParams: { 'receiver_ids.user_id': recieverId },
      },
      { status: selector ?? [] },
    )
    return Marshaler.payloadListResponse('notifications', result)
  }

  async updateNotificationStatus(recieverId, notificationIds, newStatus) {
    const result = await this._api.NotificationService.UpdateStatus(
      {
        routeParams: { 'receiver_ids.user_id': recieverId },
      },
      { ids: notificationIds, status: newStatus },
    )

    return Marshaler.payloadSingleResponse(result)
  }
}

export default Notifications
