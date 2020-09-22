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

export const deviceEvents = [
  {
    name: 'js.join.accept',
    time: '2020-04-23T09:54:48.444814226Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    correlation_ids: ['rpc:/ttn.lorawan.v3.NsJs/HandleJoin:01E6K7C4ZRQ7T9R714MA74YQWK'],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.join.cluster.success',
    time: '2020-04-23T09:54:48.445359641Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.JoinResponse',
      raw_payload: 'IKArj2vpNWEDs9AGu30NswRBasFVfTB6ZiXX1qUFjtau',
      session_keys: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
      },
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.join.cluster.attempt',
    time: '2020-04-23T09:54:48.440003103Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.JoinRequest',
      raw_payload: 'AO++rd4EAwIBBAMCAe++rd4qWFuDhTs=',
      payload: {
        m_hdr: {},
        mic: 'W4OFOw==',
        join_request_payload: {
          join_eui: '01020304DEADBEEF',
          dev_eui: 'DEADBEEF01020304',
          dev_nonce: '582A',
        },
      },
      dev_addr: '00F30390',
      selected_mac_version: '1.1.0',
      net_id: '000000',
      downlink_settings: {
        opt_neg: true,
      },
      rx_delay: 5,
      cf_list: {
        freq: [8671000, 8673000, 8675000, 8677000, 8679000],
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.join.receive',
    time: '2020-04-23T09:54:48.438409659Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'AO++rd4EAwIBBAMCAe++rd4qWFuDhTs=',
      payload: {
        m_hdr: {},
        mic: 'W4OFOw==',
        join_request_payload: {
          join_eui: '01020304DEADBEEF',
          dev_eui: 'DEADBEEF01020304',
          dev_nonce: '582A',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '868500000',
        timestamp: 2662332596,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2662332596,
          rssi: -42,
          channel_rssi: -42,
          snr: 9.5,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/ELTxv/UJ',
          channel_index: 7,
        },
      ],
      received_at: '2020-04-23T09:54:48.421622308Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
      device_channel_index: 2,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.join.process',
    time: '2020-04-23T09:54:48.627079594Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'AO++rd4EAwIBBAMCAe++rd4qWFuDhTs=',
      payload: {
        m_hdr: {},
        mic: 'W4OFOw==',
        join_request_payload: {
          join_eui: '01020304DEADBEEF',
          dev_eui: 'DEADBEEF01020304',
          dev_nonce: '582A',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '868500000',
        timestamp: 2662332596,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2662332596,
          rssi: -42,
          channel_rssi: -42,
          snr: 9.5,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/ELTxv/UJ',
          channel_index: 7,
        },
      ],
      received_at: '2020-04-23T09:54:48.421622308Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
      device_channel_index: 2,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.join.receive',
    time: '2020-04-23T09:54:48.627397191Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    correlation_ids: [
      'as:up:01E6K7C55KS9C82VWHE7Y5A80F',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.join.forward',
    time: '2020-04-23T09:54:48.629530931Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'as:up:01E6K7C55KS9C82VWHE7Y5A80F',
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
      received_at: '2020-04-23T09:54:48.627402631Z',
      join_accept: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        received_at: '2020-04-23T09:54:48.445370551Z',
      },
    },
    correlation_ids: [
      'as:up:01E6K7C55KS9C82VWHE7Y5A80F',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.join.accept.forward',
    time: '2020-04-23T09:54:48.629632783Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
      join_accept: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        received_at: '2020-04-23T09:54:48.445370551Z',
      },
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.AsNs/LinkApplication:01E6K6MXTGAAW7S2X04K71WJ8X',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.down.join.schedule.success',
    time: '2020-04-23T09:54:50.331206165Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ScheduleDownlinkResponse',
      delay: '3.089965075s',
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:downlink:01E6K7C6TT7536MW00EXGQ9BWW',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.down.join.schedule.attempt',
    time: '2020-04-23T09:54:50.330342352Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '01E96A34',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage',
      raw_payload: 'IKArj2vpNWEDs9AGu30NswRBasFVfTB6ZiXX1qUFjtau',
      request: {
        downlink_paths: [
          {
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/ELTxv/UJ',
          },
        ],
        rx1_delay: 5,
        rx1_data_rate_index: 5,
        rx1_frequency: '868500000',
        rx2_frequency: '869525000',
        priority: 'HIGHEST',
        frequency_plan_id: 'EU_863_870',
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
        'ns:downlink:01E6K7C6TT7536MW00EXGQ9BWW',
        'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
      ],
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7C4Z5QWM5BCWW2PJ04RTS',
      'ns:downlink:01E6K7C6TT7536MW00EXGQ9BWW',
      'ns:uplink:01E6K7C4Z58AGMXDD8AD5NB722',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7C4Z54KCH16AKEZVNAM74',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-04-23T09:54:53.582383525Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QJAD8wCAAAAAvyUs3UYE/1E=',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'RgT/UQ==',
        mac_payload: {
          f_hdr: {
            dev_addr: '00F30390',
            f_ctrl: {
              adr: true,
            },
          },
          frm_payload: 'vyUs3Q==',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '867900000',
        timestamp: 2667491195,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2667491195,
          rssi: -45,
          channel_rssi: -45,
          snr: 8.8,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EPve+vcJ',
          channel_index: 4,
        },
      ],
      received_at: '2020-04-23T09:54:53.580998192Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
        'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
      ],
      device_channel_index: 7,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.process',
    time: '2020-04-23T09:54:53.786333019Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QJAD8wCAAAAAvyUs3UYE/1E=',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'RgT/UQ==',
        mac_payload: {
          f_hdr: {
            dev_addr: '00F30390',
            f_ctrl: {
              adr: true,
            },
          },
          frm_payload: 'vyUs3Q==',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '867900000',
        timestamp: 2667491195,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2667491195,
          rssi: -45,
          channel_rssi: -45,
          snr: 8.8,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EPve+vcJ',
          channel_index: 4,
        },
      ],
      received_at: '2020-04-23T09:54:53.580998192Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
        'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
      ],
      device_channel_index: 7,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.mac.device_mode.indication',
    time: '2020-04-23T09:54:53.783010595Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.DeviceModeInd',
      class: 'CLASS_C',
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.class.switch.c',
    time: '2020-04-23T09:54:53.783011787Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/google.protobuf.Value',
      value: 0,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.mac.device_mode.confirmation',
    time: '2020-04-23T09:54:53.783012919Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.DeviceModeConf',
      class: 'CLASS_C',
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.mac.rekey.indication',
    time: '2020-04-23T09:54:53.783004162Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.RekeyInd',
      minor_version: 1,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.mac.rekey.confirmation',
    time: '2020-04-23T09:54:53.783008951Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.RekeyConf',
      minor_version: 1,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.receive',
    time: '2020-04-23T09:54:53.787117373Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    correlation_ids: [
      'as:up:01E6K7CA6VD8MCG3EGTPVJ8MQ2',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.forward',
    time: '2020-04-23T09:54:53.812543951Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'as:up:01E6K7CA6VD8MCG3EGTPVJ8MQ2',
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
        'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
      ],
      received_at: '2020-04-23T09:54:53.787122843Z',
      uplink_message: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        frm_payload: 'MEijVA==',
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'eui-647fdafffe007b3f',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 2667491195,
            rssi: -45,
            channel_rssi: -45,
            snr: 8.8,
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EPve+vcJ',
            channel_index: 4,
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867900000',
          timestamp: 2667491195,
        },
        received_at: '2020-04-23T09:54:53.580998192Z',
      },
    },
    correlation_ids: [
      'as:up:01E6K7CA6VD8MCG3EGTPVJ8MQ2',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.forward',
    time: '2020-04-23T09:54:53.812651543Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
        'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
      ],
      uplink_message: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        frm_payload: 'vyUs3Q==',
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'eui-647fdafffe007b3f',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 2667491195,
            rssi: -45,
            channel_rssi: -45,
            snr: 8.8,
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EPve+vcJ',
            channel_index: 4,
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867900000',
          timestamp: 2667491195,
        },
        received_at: '2020-04-23T09:54:53.580998192Z',
      },
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.AsNs/LinkApplication:01E6K6MXTGAAW7S2X04K71WJ8X',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.down.data.schedule.success',
    time: '2020-04-23T09:54:55.548895902Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ScheduleDownlinkResponse',
      delay: '3.031495689s',
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:downlink:01E6K7CBXWE88ZGDHN3ZW4MB7C',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.down.data.schedule.attempt',
    time: '2020-04-23T09:54:55.548084377Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage',
      raw_payload: 'YJAD8wCAAAAAerDWJHT79Q8=',
      request: {
        downlink_paths: [
          {
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EPve+vcJ',
          },
        ],
        rx1_delay: 5,
        rx1_data_rate_index: 5,
        rx1_frequency: '867900000',
        rx2_frequency: '869525000',
        priority: 'HIGHEST',
        frequency_plan_id: 'EU_863_870',
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
        'ns:downlink:01E6K7CBXWE88ZGDHN3ZW4MB7C',
        'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
      ],
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CA0C7B2JM9WZ8RSKW6E3',
      'ns:downlink:01E6K7CBXWE88ZGDHN3ZW4MB7C',
      'ns:uplink:01E6K7CA0C3XG47VRXH0X7BYD4',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CA0C275C5S561PEAKE29',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-04-23T09:54:58.718936620Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'gJAD8wCAAQAA8RUMuw==',
      payload: {
        m_hdr: {
          m_type: 'CONFIRMED_UP',
        },
        mic: '8RUMuw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '00F30390',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 1,
          },
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '867900000',
        timestamp: 2672632747,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2672632747,
          rssi: -42,
          channel_rssi: -42,
          snr: 10.8,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EKvHtPoJ',
          channel_index: 4,
        },
      ],
      received_at: '2020-04-23T09:54:58.718333136Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
        'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
      ],
      device_channel_index: 7,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.process',
    time: '2020-04-23T09:54:58.921802988Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'gJAD8wCAAQAA8RUMuw==',
      payload: {
        m_hdr: {
          m_type: 'CONFIRMED_UP',
        },
        mic: '8RUMuw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '00F30390',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 1,
          },
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '867900000',
        timestamp: 2672632747,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'eui-647fdafffe007b3f',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 2672632747,
          rssi: -42,
          channel_rssi: -42,
          snr: 10.8,
          uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EKvHtPoJ',
          channel_index: 4,
        },
      ],
      received_at: '2020-04-23T09:54:58.718333136Z',
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
        'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
      ],
      device_channel_index: 7,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.receive',
    time: '2020-04-23T09:54:58.922427752Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    correlation_ids: [
      'as:up:01E6K7CF7AHR5QSDV50R87TXEJ',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.forward',
    time: '2020-04-23T09:54:58.922859484Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'as:up:01E6K7CF7AHR5QSDV50R87TXEJ',
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
        'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
      ],
      received_at: '2020-04-23T09:54:58.922434335Z',
      uplink_message: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        f_cnt: 1,
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'eui-647fdafffe007b3f',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 2672632747,
            rssi: -42,
            channel_rssi: -42,
            snr: 10.8,
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EKvHtPoJ',
            channel_index: 4,
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867900000',
          timestamp: 2672632747,
        },
        received_at: '2020-04-23T09:54:58.718333136Z',
        confirmed: true,
      },
    },
    correlation_ids: [
      'as:up:01E6K7CF7AHR5QSDV50R87TXEJ',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'ns.up.data.forward',
    time: '2020-04-23T09:54:58.922953220Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
        'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
      ],
      uplink_message: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        f_cnt: 1,
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'eui-647fdafffe007b3f',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 2672632747,
            rssi: -42,
            channel_rssi: -42,
            snr: 10.8,
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EKvHtPoJ',
            channel_index: 4,
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867900000',
          timestamp: 2672632747,
        },
        received_at: '2020-04-23T09:54:58.718333136Z',
        confirmed: true,
      },
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.AsNs/LinkApplication:01E6K6MXTGAAW7S2X04K71WJ8X',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.receive',
    time: '2020-04-23T09:55:04.133547578Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    correlation_ids: [
      'as:up:01E6K7CMA5Z53A5YBDNA18H3QE',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CM3R8YDS1JV9F5EN8Y56',
      'ns:uplink:01E6K7CM3R4JY5VN4ZAN2PVFV9',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CM3R7X8EKWVWSHFZBWFF',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.forward',
    time: '2020-04-23T09:55:04.134034132Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-c',
        application_ids: {
          application_id: 'test-app2',
        },
        dev_eui: 'DEADBEEF01020304',
        join_eui: '01020304DEADBEEF',
        dev_addr: '00F30390',
      },
      correlation_ids: [
        'as:up:01E6K7CMA5Z53A5YBDNA18H3QE',
        'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
        'gs:uplink:01E6K7CM3R8YDS1JV9F5EN8Y56',
        'ns:uplink:01E6K7CM3R4JY5VN4ZAN2PVFV9',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CM3R7X8EKWVWSHFZBWFF',
      ],
      received_at: '2020-04-23T09:55:04.133553008Z',
      uplink_message: {
        session_key_id: 'AXGmdhP6Y4muAldHtH66fw==',
        f_cnt: 2,
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'eui-647fdafffe007b3f',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 2677835924,
            rssi: -40,
            channel_rssi: -40,
            snr: 7.2,
            uplink_token: 'CiIKIAoUZXVpLTY0N2ZkYWZmZmUwMDdiM2YSCGR/2v/+AHs/EJSR8vwJ',
            channel_index: 6,
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '868300000',
          timestamp: 2677835924,
        },
        received_at: '2020-04-23T09:55:03.928907755Z',
      },
    },
    correlation_ids: [
      'as:up:01E6K7CMA5Z53A5YBDNA18H3QE',
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CM3R8YDS1JV9F5EN8Y56',
      'ns:uplink:01E6K7CM3R4JY5VN4ZAN2PVFV9',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CM3R7X8EKWVWSHFZBWFF',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    name: 'as.up.data.forward',
    time: '2020-04-23T09:55:04.135034132Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'ttn-uno2',
        application_ids: {
          application_id: 'zentenes-home',
        },
        dev_eui: '0004A30B001C208A',
        join_eui: '58A0CB0001500001',
        dev_addr: '27000047',
      },
      correlation_ids: [
        'as:up:01E6XXPP8H76JPT23PDMQNNX3B',
        'gs:conn:01E6VZ6NHWEN93DW41TK7ZYZ9Z',
        'gs:uplink:01E6XXPP1X4DV7NE9929RD02WA',
        'ns:uplink:01E6XXPP1Z0B7NPG676V8YSXGE',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6XXPP1YPNCDG5AT0D7FCDDE',
      ],
      received_at: '2020-04-27T13:37:26.802452748Z',
      uplink_message: {
        session_key_id: 'AXG7uL9WnE8oFivIfU7fCw==',
        f_port: 1,
        f_cnt: 18,
        frm_payload: 'AQ==',
        decoded_payload: [22.3, 'ON'],
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'bafonins-ttig',
              eui: '58A0CBFFFE8010D6',
            },
            time: '2020-04-27T13:37:26.957297086Z',
            timestamp: 1110787324,
            rssi: -33,
            channel_rssi: -33,
            snr: 6.25,
            location: {
              latitude: 56.961865,
              longitude: 24.003738,
              altitude: 1,
              source: 'SOURCE_REGISTRY',
            },
            uplink_token: 'ChsKGQoNYmFmb25pbnMtdHRpZxIIWKDL//6AENYQ/InVkQQ=',
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867500000',
          timestamp: 1110787324,
          time: '2020-04-27T13:37:26.957297086Z',
        },
        received_at: '2020-04-27T13:37:26.591056488Z',
      },
    },
  },
  {
    name: 'as.up.data.forward',
    time: '2020-04-23T09:55:04.134034142Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'ttn-uno2',
        application_ids: {
          application_id: 'zentenes-home',
        },
        dev_eui: '0004A30B001C208A',
        join_eui: '58A0CB0001500001',
        dev_addr: '27000047',
      },
      correlation_ids: [
        'as:up:01E6XXPP8H76JPT23PDMQNNX3B',
        'gs:conn:01E6VZ6NHWEN93DW41TK7ZYZ9Z',
        'gs:uplink:01E6XXPP1X4DV7NE9929RD02WA',
        'ns:uplink:01E6XXPP1Z0B7NPG676V8YSXGE',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6XXPP1YPNCDG5AT0D7FCDDE',
      ],
      received_at: '2020-04-27T13:37:26.802452748Z',
      uplink_message: {
        session_key_id: 'AXEGzcGdYrEztqzzpAaNYg==',
        f_port: 102,
        f_cnt: 5271,
        frm_payload: 'Afs2AADArgA=',
        decoded_payload: {
          battery: 100,
          events: 'motion',
          status: 1,
          temp: 22.3,
          voltage: 3.6,
        },
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'bafonins-ttig',
              eui: '58A0CBFFFE8010D6',
            },
            time: '2020-04-27T13:37:26.957297086Z',
            timestamp: 1110787324,
            rssi: -33,
            channel_rssi: -33,
            snr: 6.25,
            location: {
              latitude: 56.961865,
              longitude: 24.003738,
              altitude: 1,
              source: 'SOURCE_REGISTRY',
            },
            uplink_token: 'ChsKGQoNYmFmb25pbnMtdHRpZxIIWKDL//6AENYQ/InVkQQ=',
          },
        ],
        settings: {
          data_rate: {
            lora: {
              bandwidth: 125000,
              spreading_factor: 7,
            },
          },
          data_rate_index: 5,
          coding_rate: '4/5',
          frequency: '867500000',
          timestamp: 1110787324,
          time: '2020-04-27T13:37:26.957297086Z',
        },
        received_at: '2020-04-27T13:37:26.591056488Z',
      },
    },
  },
  {
    name: 'ns.down.data.schedule.success',
    time: '2020-04-23T09:55:00.665284566Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/gatewayserver',
      name: 'host_handle',
      message_format: 'host `{host}` failed to handle message',
      attributes: {
        host: 'cluster',
      },
      cause: {
        namespace: 'pkg/networkserver',
        name: 'device_not_found',
        message_format: 'device not found',
        correlation_id: 'df971dc6e7c5402596576816401ade98',
        code: 5,
      },
      code: 5,
    },
    correlation_ids: [
      'gs:conn:01E6K64KJPJRE9CNM6KGDP7N3R',
      'gs:uplink:01E6K7CF0XNG6Y965C4WF76CVR',
      'ns:downlink:01E6K7CGXRGHCG2P2VB66MRE1X',
      'ns:uplink:01E6K7CF0Y27X81HZZ5AN8D59F',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E6K7CF0Y3NP9RY89DGJ74XJE',
    ],
    origin: 'cobalt',
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
  },
  {
    time: '2020-04-23T09:55:00.665284567Z',
    name: 'as.down.data.forward',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationDownlink',
      f_port: 1,
      frm_payload: '6H6t',
      correlation_ids: [
        'as:downlink:01E70VK078A64W1DY6NEQ5S8KE',
        'rpc:/ttn.lorawan.v3.AppAs/DownlinkQueueReplace:01E70VK05K9ZSV24QEBS486JXF',
      ],
    },
  },
  {
    time: '2020-04-23T09:55:00.665284568Z',
    name: 'as.down.data.receive',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-c',
          application_ids: {
            application_id: 'test-app2',
          },
          dev_eui: 'DEADBEEF01020304',
          join_eui: '01020304DEADBEEF',
          dev_addr: '00F30390',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationDownlink',
      f_port: 1,
      frm_payload: '6H6t',
      correlation_ids: [
        'as:downlink:01E70VK078A64W1DY6NEQ5S8KE',
        'rpc:/ttn.lorawan.v3.AppAs/DownlinkQueueReplace:01E70VK05K9ZSV24QEBS486JXF',
      ],
    },
  },
].reverse()

export const events = [
  ...deviceEvents,
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.create',
    time: '2020-04-23T09:39:04.134034132Z',
  },
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.delete',
    time: '2020-04-23T09:40:04.134034132Z',
  },
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.update',
    time: '2020-04-23T09:41:04.134034132Z',
  },
]

export const gatewayEvents = [
  {
    name: 'gs.up.receive',
    time: '2020-04-23T09:55:00.665284569Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QJkAACcAAAABUbmOVyg=',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'uY5XKA==',
        mac_payload: {
          f_hdr: {
            dev_addr: '27000099',
            f_ctrl: {},
          },
          f_port: 1,
          frm_payload: 'UQ==',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        coding_rate: '4/5',
        frequency: '867500000',
        timestamp: 850105076,
        time: '2020-05-02T15:33:15.075889110Z',
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'bafonins-ttig',
            eui: '58A0CBFFFE8010D6',
          },
          time: '2020-05-02T15:33:15.075889110Z',
          timestamp: 850105076,
          rssi: -41,
          channel_rssi: -41,
          snr: 6.75,
          uplink_token: 'ChsKGQoNYmFmb25pbnMtdHRpZxIIWKDL//6AENYQ9KWulQM=',
        },
      ],
      received_at: '2020-05-02T15:33:14.610279766Z',
      correlation_ids: [
        'gs:conn:01E78HNYZTNGMXNA3S2GWS80T6',
        'gs:uplink:01E7B0AA7JDMX0MQZJ82PQ1VR2',
      ],
    },
  },
  {
    name: 'gs.up.forward',
    time: '2020-04-23T09:55:01.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
  },
  {
    name: 'gs.up.drop',
    time: '2020-04-23T09:55:02.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/gatewayserver',
      name: 'host_handle',
      message_format: 'host `{host}` failed to handle message',
      attributes: {
        host: 'cluster',
      },
      cause: {
        namespace: 'pkg/networkserver',
        name: 'device_not_found',
        message_format: 'device not found',
        correlation_id: '1ed3053673dc4e62adb3fd82eec9386a',
        code: 5,
      },
      code: 5,
    },
  },
  {
    name: 'gs.up.receive',
    time: '2020-04-23T09:55:03.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'AAEAUAEAy6BYiiAcAAujBABBPzKr7WA=',
      payload: {
        m_hdr: {},
        mic: 'MqvtYA==',
        join_request_payload: {
          join_eui: '58A0CB0001500001',
          dev_eui: '0004A30B001C208A',
          dev_nonce: '3F41',
        },
      },
      settings: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        coding_rate: '4/5',
        frequency: '868100000',
        timestamp: 851245003,
        time: '2020-05-02T15:33:16.212656021Z',
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'bafonins-ttig',
            eui: '58A0CBFFFE8010D6',
          },
          time: '2020-05-02T15:33:16.212656021Z',
          timestamp: 851245003,
          rssi: -41,
          channel_rssi: -41,
          snr: 10,
          uplink_token: 'ChsKGQoNYmFmb25pbnMtdHRpZxIIWKDL//6AENYQy+/zlQM=',
        },
      ],
      received_at: '2020-05-02T15:33:15.745385355Z',
      correlation_ids: [
        'gs:conn:01E78HNYZTNGMXNA3S2GWS80T6',
        'gs:uplink:01E7B0ABB1N9GY31G4TCZJY21E',
      ],
    },
  },
  {
    name: 'gs.up.forward',
    time: '2020-04-23T09:55:04.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
  },
  {
    name: 'gs.up.forward',
    time: '2020-04-23T09:55:05.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
  },
  {
    name: 'gs.down.send',
    time: '2020-04-23T09:55:06.665284565Z',
    identifiers: [
      {
        gateway_ids: { gateway_id: 'test-gtw-id' },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage',
      raw_payload: 'IMcelqpjPifS7/l1Bs5Gs4vEBZdoikcTpumceVmXRdtj',
      scheduled: {
        data_rate: {
          lora: {
            bandwidth: 125000,
            spreading_factor: 7,
          },
        },
        data_rate_index: 5,
        coding_rate: '4/5',
        frequency: '868100000',
        timestamp: 856245003,
        downlink: {
          tx_power: 16.15,
          invert_polarization: true,
        },
      },
      correlation_ids: [
        'gs:conn:01E78HNYZTNGMXNA3S2GWS80T6',
        'gs:uplink:01E7B0ABB1N9GY31G4TCZJY21E',
        'ns:downlink:01E7B0AD8P52G8HFKJ2MQXAYEG',
        'ns:uplink:01E7B0ABB34Z44013JNAAEFMCR',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E7B0ABB306JH68DEMNMXBEF0',
        'gs:conn:01E78HNYZTNGMXNA3S2GWS80T6',
        'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01E7B0AD8PCHHDPETTJB6R3T7P',
      ],
    },
  },
].reverse()

export const organizationEvents = [
  {
    identifiers: [{ organization_ids: { organization_id: 'test-org-id' } }],
    name: 'organization.api-key.create',
    time: '2020-04-23T09:39:04.134034132Z',
  },
  {
    identifiers: [{ organization_ids: { organization_id: 'test-org-id' } }],
    name: 'organization.update',
    time: '2020-04-23T09:41:04.134034132Z',
  },
]
