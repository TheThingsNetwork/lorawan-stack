// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

export const crud = /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+(create|delete|update)$/
export const crudCreate = /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+create$/
export const crudDelete = /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+delete$/
export const crudUpdate = /^(ns\.|as\.|js\.)?([a-z0-9](?:[-_]?[a-z0-9]){2,}\.)+update$/
export const deviceUplink = /^(ns|as)\.up(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
export const deviceDownlink = /^(ns|as)\.down(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
export const deviceJoin = /^(js|ns|as)(\.up|\.down)?\.(join|rejoin)(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
export const gatewayUplink = /^gs\.up(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
export const gatewayDownlink = /^gs\.down(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
export const gatewayConnection = /^gs.gateway.(connect|disconnect)$/
export const ns = /^ns(\.[a-z0-9](?:[-_]?[a-z0-9]){2,})+$/
