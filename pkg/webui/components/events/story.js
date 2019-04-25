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

import React from 'react'
import bind from 'autobind-decorator'
import { storiesOf } from '@storybook/react'

import Events from '.'

const events = [
  { identifiers: [{ application_ids: { application_id: 'admin-app' }}], name: 'application.create', time: '2019-03-28T13:18:13.376022Z' },
  { identifiers: [{ application_ids: { application_id: 'admin-app' }}], name: 'application.delete', time: '2019-03-28T13:18:24.376022Z' },
  { identifiers: [{ application_ids: { application_id: 'admin-app' }}], name: 'application.update', time: '2019-03-28T13:18:39.376022Z' },
  { identifiers: [{ application_ids: { application_id: 'admin-app' }}], name: 'application.delete', time: '2019-03-28T13:18:48.376022Z' },
  { name: 'gs.gateway.connect', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw' }}], correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'gs.up.receive', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage', raw_payload: 'gNik0AAAAAAB6LgQBaNT', payload: { m_hdr: { m_type: 'CONFIRMED_UP' }, mic: 'EAWjUw==', mac_payload: { f_hdr: { dev_addr: '00D0A4D8', f_ctrl: {}}, f_port: 1, frm_payload: '6Lg=' }}, settings: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 12 }}, coding_rate: '4/5', frequency: '868100000', timestamp: 909813543, time: '2019-04-15T09:23:56.844839Z' }, rx_metadata: [{ gateway_ids: { gateway_id: 'admin-gtw' }, time: '2019-04-15T09:23:56.844839Z', timestamp: 909813543, uplink_token: 'ChcKFQoJYWRtaW4tZ3R3EggBAgMEBQYHCBCnzuqxAw==' }], received_at: '2019-04-15T09:23:56.994729Z', correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ]}, correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'ns.up.merge_metadata', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/google.protobuf.Value', value: 1 }, correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'as.up.data.receive', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], correlation_ids: [ 'as:up:01D8G4YV4SAPR3XFJGTA7HBHSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'ns.up.data.forward', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'as.up.data.forward', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp', end_device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }, correlation_ids: [ 'as:up:01D8G4YV4SAPR3XFJGTA7HBHSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], received_at: '2019-04-15T09:23:56.996615Z', uplink_message: { session_key_id: 'AWogTGghnCfSJwgfSwASXQ==', f_port: 1, frm_payload: 'q80=', rx_metadata: [{ gateway_ids: { gateway_id: 'admin-gtw' }, time: '2019-04-15T09:23:56.844839Z', timestamp: 909813543, uplink_token: 'ChcKFQoJYWRtaW4tZ3R3EggBAgMEBQYHCBCnzuqxAw==' }], settings: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 12 }}, coding_rate: '4/5', frequency: '868100000', timestamp: 909813543, time: '2019-04-15T09:23:56.844839Z' }}}, correlation_ids: [ 'as:up:01D8G4YV4SAPR3XFJGTA7HBHSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'as:up:01D8G4YV4SAPR3XFJGTA7HBHSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'ns.mac.dev_status.request', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'gs.down.send', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage', raw_payload: 'YNik0AAgAQAAEX2avOs=', scheduled: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 12 }}, coding_rate: '4/5', frequency: '868100000', timestamp: 914813543, downlink: { tx_power: 16.15, invert_polarization: true }}, correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:downlink:01D8G4YV66Y46G6TERJHX6Y83T', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4YV66Y9CT36Y034RCETAN' ]}, correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4YV66Y9CT36Y034RCETAN', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'gs:uplink:01D8G4YTY2SSY75ZJXWRA8S216', 'ns:downlink:01D8G4YV66Y46G6TERJHX6Y83T', 'ns:uplink:01D8G4YTY4C61NYHTP2D3E96SM', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4YTY4R56G9BJ31XGCBC0K', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4YV66Y9CT36Y034RCETAN' ], origin: 'htdvissermbp' },
  { name: 'gs.gateway.disconnect', time: '2019-03-28T13:18:48.376022Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], correlation_ids: [ 'gs:conn:01D8G4YTWF26G9DZ7ES7KCND48', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4YTSKD7DZFMXJBMCK6CSD' ], origin: 'htdvissermbp' },
  { name: 'gs.gateway.connect', time: '2019-04-15T09:20:39.435488Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw' }}], correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'gs.up.receive', time: '2019-04-15T09:20:39.449267Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage', raw_payload: 'AAgHBgUEAwIBCAcGBQQDAgEwAQBPZ7E=', payload: { m_hdr: {}, mic: 'AE9nsQ==', join_request_payload: { join_eui: '0102030405060708', dev_eui: '0102030405060708', dev_nonce: '0130' }}, settings: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 12 }}, coding_rate: '4/5', frequency: '868100000', timestamp: 712249018, time: '2019-04-15T09:20:39.280314Z' }, rx_metadata: [{ gateway_ids: { gateway_id: 'admin-gtw' }, time: '2019-04-15T09:20:39.280314Z', timestamp: 712249018, uplink_token: 'ChcKFQoJYWRtaW4tZ3R3EggBAgMEBQYHCBC6ndDTAg==' }], received_at: '2019-04-15T09:20:39.436603Z', correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ]}, correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'js.join.accept', time: '2019-04-15T09:20:39.463256Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708' }}], correlation_ids: [ 'rpc:/ttn.lorawan.v3.NsJs/HandleJoin:01D8G4RT0Y8X66XY44T5JW22ZC' ], origin: 'htdvissermbp' },
  { name: 'ns.up.join.forward', time: '2019-04-15T09:20:39.464551Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708' }}], correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'ns.up.merge_metadata', time: '2019-04-15T09:20:39.654493Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/google.protobuf.Value', value: 1 }, correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'as.up.join.receive', time: '2019-04-15T09:20:39.661194Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], correlation_ids: [ 'as:up:01D8G4RT7DNTF5QT9J0NE4SFJ8', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'as.up.join.forward', time: '2019-04-15T09:20:39.664860Z', identifiers: [{ device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp', end_device_ids: { device_id: 'admin-dev-otaa', application_ids: { application_id: 'admin-app' }, dev_eui: '0102030405060708', join_eui: '0102030405060708', dev_addr: '00D0A4D8' }, correlation_ids: [ 'as:up:01D8G4RT7DNTF5QT9J0NE4SFJ8', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], received_at: '2019-04-15T09:20:39.451100Z', join_accept: { session_key_id: 'AWogTGghnCfSJwgfSwASXQ==' }}, correlation_ids: [ 'as:up:01D8G4RT7DNTF5QT9J0NE4SFJ8', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'as:up:01D8G4RT7DNTF5QT9J0NE4SFJ8', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
  { name: 'gs.down.send', time: '2019-04-15T09:20:39.672510Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], data: { '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage', raw_payload: 'IJJfgSo3PVroy3IL9z9MUwojoT9Yi0CtUaRv1xdIXIiE', scheduled: { data_rate: { lora: { bandwidth: 125000, spreading_factor: 12 }}, coding_rate: '4/5', frequency: '868100000', timestamp: 717249018, downlink: { tx_power: 16.15, invert_polarization: true }}, correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:downlink:01D8G4RT7QHVG8G489ES6K2QJD', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4RT7RR7D5FSZGV7YTVMEY' ]}, correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4RT7RR7D5FSZGV7YTVMEY', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'gs:uplink:01D8G4RT0SJK80Y33195AP8YK4', 'ns:downlink:01D8G4RT7QHVG8G489ES6K2QJD', 'ns:uplink:01D8G4RT0VVYWXB188X8FXJEV8', 'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D8G4RT0T74QP3GXGH3G0KGDA', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK', 'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D8G4RT7RR7D5FSZGV7YTVMEY' ], origin: 'htdvissermbp' },
  { name: 'gs.gateway.disconnect', time: '2019-04-15T09:20:39.679935Z', identifiers: [{ gateway_ids: { gateway_id: 'admin-gtw', eui: '0102030405060708' }}], correlation_ids: [ 'gs:conn:01D8G4RSYNH95CEQ0CY3Y6ESH8', 'rpc:/ttn.lorawan.v3.GtwGs/LinkGateway:01D8G4RSVNGBNQMGZ3Q1QJQFKK' ], origin: 'htdvissermbp' },
]

const getRandomEvent = function () {
  const rndIndex = Math.floor(Math.random() * events.length)
  const event = events[rndIndex]

  return { ...event, time: new Date().toISOString() }
}

@bind
class Example extends React.Component {

  constructor (props) {
    super(props)

    this.state = {
      paused: false,
    }
  }

  onPause () {
    this.setState(prev => ({ paused: !prev.paused }))
  }

  onClear () {
    const { onClear } = this.props

    if (onClear) {
      onClear()
    }
  }

  render () {
    const { paused } = this.state
    const { events } = this.props

    return (
      <Events
        emitterId="my-app"
        events={events}
        paused={paused}
        onClear={this.onClear}
        onPause={this.onPause}
      />
    )
  }
}

@bind
class DynamicExample extends React.Component {

  constructor (props) {
    super(props)

    this.state = {
      paused: false,
      emittedEvents: [],
    }

    const self = this
    this.interval = setInterval(
      function () {
        if (!self.state.paused) {
          self.setState(function (prev) {
            const events = prev.emittedEvents
            const size = events.length
            if (prev.emittedEvents.length >= 100) {
              return { emittedEvents: [ getRandomEvent(), ...events.slice(0, size - 1) ]}
            }

            return { emittedEvents: [ getRandomEvent(), ...prev.emittedEvents ]}
          })
        }
      }, 2000)
  }

  onPause () {
    this.setState(prev => ({ paused: !prev.paused }))
  }

  onClear () {
    const { onClear } = this.props

    if (onClear) {
      onClear()
    }
  }

  componentWillUnmount () {
    clearInterval(this.interval)
  }

  render () {
    const { emittedEvents, paused } = this.state

    return (
      <Events
        emitterId="my-app"
        events={emittedEvents}
        paused={paused}
        onClear={this.onClear}
        onPause={this.onPause}
      />
    )
  }
}

storiesOf('Events/Full', module)
  .add('Empty', () => (
    <Example events={[]} />
  ))
  .add('Default', () => (
    <Example events={events} />
  ))
  .add('Dynamic', () => (
    <DynamicExample />
  ))
