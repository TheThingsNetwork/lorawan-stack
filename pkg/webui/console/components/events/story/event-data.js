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
    name: 'ns.mac.link_adr.request',
    time: '2020-09-25T13:46:54.243812282Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.LinkADRReq',
      data_rate_index: 5,
      tx_power_index: 1,
      channel_mask: [
        true,
        true,
        true,
        true,
        true,
        true,
        true,
        true,
        false,
        false,
        false,
        false,
        false,
        false,
        false,
        false,
      ],
      nb_trans: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:downlink:01EK2R8HCWTRVJVVP4HPYV86NN',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8HD3GF7385AXR4028NJE',
  },
  {
    name: 'ns.down.data.schedule.success',
    time: '2020-09-25T13:46:54.243805642Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ScheduleDownlinkResponse',
      delay: '2.813388578s',
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:downlink:01EK2R8HCWTRVJVVP4HPYV86NN',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8HD3ZRBZXV3S8MFATKKV',
  },
  {
    name: 'ns.down.data.schedule.attempt',
    time: '2020-09-25T13:46:54.236832035Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage',
      raw_payload: 'YAsAACeFAgADUf8AAWRrIjc=',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_DOWN',
        },
        mic: 'ZGsiNw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 2,
            f_opts: 'A1H/AAE=',
          },
          full_f_cnt: 2,
        },
      },
      request: {
        downlink_paths: [
          {
            uplink_token:
              'CiMKIQoVcm9tYW4ta29uYS1taWNyby1ob21lEghkf9r//gB7PxCsw4NxGgsIzOm3+wUQ6ZioGSDgj8aD84MB',
          },
        ],
        rx1_delay: 5,
        rx1_data_rate_index: 5,
        rx1_frequency: '868300000',
        rx2_data_rate_index: 3,
        rx2_frequency: '869525000',
        priority: 'HIGHEST',
        frequency_plan_id: 'EU_863_870_TTN',
      },
      correlation_ids: [
        'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
        'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
        'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
        'ns:downlink:01EK2R8HCWTRVJVVP4HPYV86NN',
        'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
      ],
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:downlink:01EK2R8HCWTRVJVVP4HPYV86NN',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8HCWT925RG5YMBCRFBCB',
  },
  {
    name: 'ns.up.data.forward',
    time: '2020-09-25T13:46:53.529570554Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-01',
        application_ids: {
          application_id: 'tti-ch-test-app',
        },
        dev_eui: '0004A30B001C1E48',
        join_eui: '8000000000000003',
        dev_addr: '2700000B',
      },
      correlation_ids: [
        'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
        'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
        'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
        'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
      ],
      uplink_message: {
        session_key_id: 'AXTFUoae2WV6TtuOUy0bHQ==',
        f_port: 1,
        f_cnt: 202,
        frm_payload: '1A==',
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'test-gateway-01',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 237035948,
            rssi: -49,
            channel_rssi: -49,
            snr: 6.8,
            uplink_token:
              'CiMKIQoVcm9tYW4ta29uYS1taWNyby1ob21lEghkf9r//gB7PxCsw4NxGgsIzOm3+wUQ6ZioGSDgj8aD84MB',
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
          timestamp: 237035948,
        },
        received_at: '2020-09-25T13:46:52.433197574Z',
      },
    },
    correlation_ids: ['rpc:/ttn.lorawan.v3.AsNs/LinkApplication:01EK07JNQCW64QV8JXHBVKN4JW'],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8GPSB7WR2GNW7TP6A717',
  },
  {
    name: 'as.up.data.forward',
    time: '2020-09-25T13:46:53.445120375Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-01',
        application_ids: {
          application_id: 'tti-ch-test-app',
        },
        dev_eui: '0004A30B001C1E48',
        join_eui: '8000000000000003',
        dev_addr: '2700000B',
      },
      correlation_ids: [
        'as:up:01EK2R8GM3JN90QF338G3G2NS0',
        'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
        'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
        'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
        'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
      ],
      received_at: '2020-09-25T13:46:53.444154877Z',
      uplink_message: {
        session_key_id: 'AXTFUoae2WV6TtuOUy0bHQ==',
        f_port: 1,
        f_cnt: 202,
        frm_payload: 'AQ==',
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'test-gateway-01',
              eui: '647FDAFFFE007B3F',
            },
            timestamp: 237035948,
            rssi: -49,
            channel_rssi: -49,
            snr: 6.8,
            uplink_token:
              'CiMKIQoVcm9tYW4ta29uYS1taWNyby1ob21lEghkf9r//gB7PxCsw4NxGgsIzOm3+wUQ6ZioGSDgj8aD84MB',
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
          timestamp: 237035948,
        },
        received_at: '2020-09-25T13:46:52.433197574Z',
      },
    },
    correlation_ids: [
      'as:up:01EK2R8GM3JN90QF338G3G2NS0',
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-6-7.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8GM5XZBD1S5PGPRCE5NV',
  },
  {
    name: 'as.up.data.receive',
    time: '2020-09-25T13:46:53.443598332Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    correlation_ids: [
      'as:up:01EK2R8GM3JN90QF338G3G2NS0',
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-6-7.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8GM30TTJFBQP763TC3Z9',
  },
  {
    name: 'ns.up.data.process',
    time: '2020-09-25T13:46:53.439670065Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QAsAACeCygADAwHUcPJPkw==',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'cPJPkw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 202,
            f_opts: 'AwM=',
          },
          f_port: 1,
          frm_payload: '1A==',
          full_f_cnt: 202,
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
        frequency: '868300000',
        timestamp: 237035948,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'test-gateway-01',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 237035948,
          rssi: -49,
          channel_rssi: -49,
          snr: 6.8,
          uplink_token:
            'CiMKIQoVcm9tYW4ta29uYS1taWNyby1ob21lEghkf9r//gB7PxCsw4NxGgsIzOm3+wUQ6ZioGSDgj8aD84MB',
          channel_index: 6,
        },
      ],
      received_at: '2020-09-25T13:46:52.433197574Z',
      correlation_ids: [
        'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
        'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
        'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
        'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
      ],
      device_channel_index: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8GKZGRA9YZHKDQHMVG45',
  },
  {
    name: 'ns.mac.link_adr.answer.reject',
    time: '2020-09-25T13:46:52.934976037Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.MACCommand.LinkADRAns',
      channel_mask_ack: true,
      data_rate_index_ack: true,
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8G465XN09AC629G9JRF7',
  },
  {
    name: 'ns.up.data.drop',
    time: '2020-09-25T13:46:52.832586045Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/networkserver',
      name: 'duplicate',
      message_format: 'uplink is a duplicate',
      code: 9,
    },
    correlation_ids: [
      'gs:conn:01EK2NJ6C2BPB5G08BE10AZCCB',
      'gs:up:host:01EK2NJ6CGSPES1V0FCCAK8M9B',
      'gs:uplink:01EK2R8F9EWS31N57RZXRQE3X5',
      'ns:uplink:01EK2R8FXZH9RCHDTZE3EVKHFX',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXZMQDZB1SXGAY2C8N1',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8G10E3YQDDWNY2B9TZ36',
  },
  {
    name: 'ns.up.data.drop',
    time: '2020-09-25T13:46:52.831971251Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/networkserver',
      name: 'duplicate',
      message_format: 'uplink is a duplicate',
      code: 9,
    },
    correlation_ids: [
      'gs:conn:01EK2KVSDQDR4MMSR4KVK7RJ3D',
      'gs:up:host:01EK2KVSE2VFZW12BBAD20YJWY',
      'gs:uplink:01EK2R8F95SPS4PS195NFEMBQ5',
      'ns:uplink:01EK2R8FXYMZBQ9PPWYZWY61EW',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXYFZ6YATW7KVMSDEEM',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8G0ZBEM99M6AKT65AEKN',
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-09-25T13:46:52.830998763Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QAsAACeCygADAwHUcPJPkw==',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'cPJPkw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 202,
            f_opts: 'AwM=',
          },
          f_port: 1,
          frm_payload: '1A==',
          full_f_cnt: 202,
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
        frequency: '868300000',
        timestamp: 2826024707,
        time: '2020-09-25T13:46:52.049526929Z',
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'test-gateway-02',
            eui: '58A0CBFFFE800568',
          },
          time: '2020-09-25T13:46:52.049526929Z',
          timestamp: 2826024707,
          rssi: -22,
          channel_rssi: -22,
          snr: 9.75,
          uplink_token:
            'CiEKHwoTcm9tYW4tdHRpZy0yMDE5LXR0YxIIWKDL//6ABWgQg+7GwwoaCwjM6bf7BRD93aAlILjHy+GfUg==',
        },
      ],
      received_at: '2020-09-25T13:46:52.735412657Z',
      correlation_ids: [
        'gs:conn:01EK2NJ6C2BPB5G08BE10AZCCB',
        'gs:up:host:01EK2NJ6CGSPES1V0FCCAK8M9B',
        'gs:uplink:01EK2R8F9EWS31N57RZXRQE3X5',
        'ns:uplink:01EK2R8FXZH9RCHDTZE3EVKHFX',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXZMQDZB1SXGAY2C8N1',
      ],
      device_channel_index: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2NJ6C2BPB5G08BE10AZCCB',
      'gs:up:host:01EK2NJ6CGSPES1V0FCCAK8M9B',
      'gs:uplink:01EK2R8F9EWS31N57RZXRQE3X5',
      'ns:uplink:01EK2R8FXZH9RCHDTZE3EVKHFX',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXZMQDZB1SXGAY2C8N1',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8G0YKR37MSHGA7RX4CD2',
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-09-25T13:46:52.830223477Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QAsAACeCygADAwHUcPJPkw==',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'cPJPkw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 202,
            f_opts: 'AwM=',
          },
          f_port: 1,
          frm_payload: '1A==',
          full_f_cnt: 202,
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
        frequency: '868300000',
        timestamp: 313787235,
        time: '2020-09-25T13:46:52.024125099Z',
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'test-gateway-03',
            eui: '58A0CBFFFE801244',
          },
          time: '2020-09-25T13:46:52.024125099Z',
          timestamp: 313787235,
          rssi: -60,
          channel_rssi: -60,
          snr: 9.25,
          uplink_token:
            'CiEKHwoTcm9tYW4tdHRpZy0yMDIwLXR0YxIIWKDL//6AEkQQ44bQlQEaCwjM6bf7BRC4uachILj1tPmQhgE=',
        },
      ],
      received_at: '2020-09-25T13:46:52.734717502Z',
      correlation_ids: [
        'gs:conn:01EK2KVSDQDR4MMSR4KVK7RJ3D',
        'gs:up:host:01EK2KVSE2VFZW12BBAD20YJWY',
        'gs:uplink:01EK2R8F95SPS4PS195NFEMBQ5',
        'ns:uplink:01EK2R8FXYMZBQ9PPWYZWY61EW',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXYFZ6YATW7KVMSDEEM',
      ],
      device_channel_index: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2KVSDQDR4MMSR4KVK7RJ3D',
      'gs:up:host:01EK2KVSE2VFZW12BBAD20YJWY',
      'gs:uplink:01EK2R8F95SPS4PS195NFEMBQ5',
      'ns:uplink:01EK2R8FXYMZBQ9PPWYZWY61EW',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FXYFZ6YATW7KVMSDEEM',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8G0YZQ8XCZTX8K1D9S2Y',
  },
  {
    name: 'ns.up.data.drop',
    time: '2020-09-25T13:46:52.737792936Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/networkserver',
      name: 'duplicate',
      message_format: 'uplink is a duplicate',
      code: 9,
    },
    correlation_ids: [
      'gs:conn:01EK2M1PP8JNQTQSG2Q9EBRCKM',
      'gs:up:host:01EK2M1PPF1NZF1W5KW76EZH4Z',
      'gs:uplink:01EK2R8F8J9VB6F3Q0ZN4PERP9',
      'ns:uplink:01EK2R8FTRRTCSCTE1HBZPNF88',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FTR347C5GJKAQMJ7SJR',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8FY1RBDAE03PER5NWTT9',
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-09-25T13:46:52.736255123Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QAsAACeCygADAwHUcPJPkw==',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'cPJPkw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 202,
            f_opts: 'AwM=',
          },
          f_port: 1,
          frm_payload: '1A==',
          full_f_cnt: 202,
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
        frequency: '868300000',
        timestamp: 125013195,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'test-gateway-04',
            eui: '647FDAFFFE007C01',
          },
          timestamp: 125013195,
          rssi: -38,
          channel_rssi: -38,
          snr: 10.2,
          uplink_token:
            'CiIKIAoUcm9tYW4ta29uYS1taWNyby1kZXYSCGR/2v/+AHwBEMuZzjsaCwjM6bf7BRDnw4kYIPjx99rRgAE=',
          channel_index: 1,
        },
      ],
      received_at: '2020-09-25T13:46:52.632704943Z',
      correlation_ids: [
        'gs:conn:01EK2M1PP8JNQTQSG2Q9EBRCKM',
        'gs:up:host:01EK2M1PPF1NZF1W5KW76EZH4Z',
        'gs:uplink:01EK2R8F8J9VB6F3Q0ZN4PERP9',
        'ns:uplink:01EK2R8FTRRTCSCTE1HBZPNF88',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FTR347C5GJKAQMJ7SJR',
      ],
      device_channel_index: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2M1PP8JNQTQSG2Q9EBRCKM',
      'gs:up:host:01EK2M1PPF1NZF1W5KW76EZH4Z',
      'gs:uplink:01EK2R8F8J9VB6F3Q0ZN4PERP9',
      'ns:uplink:01EK2R8FTRRTCSCTE1HBZPNF88',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FTR347C5GJKAQMJ7SJR',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8FY08B831X3PH1AYFJVM',
  },
  {
    name: 'ns.up.data.receive',
    time: '2020-09-25T13:46:52.533517058Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1E48',
          join_eui: '8000000000000003',
          dev_addr: '2700000B',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QAsAACeCygADAwHUcPJPkw==',
      payload: {
        m_hdr: {
          m_type: 'CONFIRMED_UP',
        },
        mic: 'cPJPkw==',
        mac_payload: {
          f_hdr: {
            dev_addr: '2700000B',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 202,
            f_opts: 'AwM=',
          },
          f_port: 1,
          frm_payload: '1A==',
          full_f_cnt: 202,
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
        frequency: '868300000',
        timestamp: 237035948,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'test-gateway-01',
            eui: '647FDAFFFE007B3F',
          },
          timestamp: 237035948,
          rssi: -49,
          channel_rssi: -49,
          snr: 6.8,
          uplink_token:
            'CiMKIQoVcm9tYW4ta29uYS1taWNyby1ob21lEghkf9r//gB7PxCsw4NxGgsIzOm3+wUQ6ZioGSDgj8aD84MB',
          channel_index: 6,
        },
      ],
      received_at: '2020-09-25T13:46:52.433197574Z',
      correlation_ids: [
        'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
        'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
        'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
        'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
      ],
      device_channel_index: 1,
    },
    correlation_ids: [
      'gs:conn:01EK2KY9B4Q2Y6QT1GAEEB2TQX',
      'gs:up:host:01EK2KY9BB88KBZFGQCQYTGYPK',
      'gs:uplink:01EK2R8F8NSVXWK1YRC42DS0D0',
      'ns:uplink:01EK2R8FMHS2K1B5FPBE65PFQN',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01EK2R8FMGZGHJ1X6K6QQ94JX8',
    ],
    origin: 'ip-10-20-12-205.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01EK2R8FQNEB3B8SFNH0G68KH3',
  },
  {
    name: 'as.up.data.forward',
    time: '2021-07-07T10:50:00.680482189Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
        },
      },
      {
        device_ids: {
          device_id: 'test-dev-01',
          application_ids: {
            application_id: 'tti-ch-test-app',
          },
          dev_eui: '0004A30B001C1452',
          join_eui: '800000000000000C',
          dev_addr: '26001B76',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationUp',
      end_device_ids: {
        device_id: 'test-dev-01',
        application_ids: {
          application_id: 'tti-ch-test-app',
        },
        dev_eui: '0004A30B001C1452',
        join_eui: '800000000000000C',
        dev_addr: '26001B76',
      },
      correlation_ids: [
        'as:up:01FA09DFHJ0FSWAPWDD3E1RDG6',
        'gs:conn:01FA04C0D25ZM679FXF7QWM6KP',
        'gs:up:host:01FA04C0DNXVBNVWGSE75FVSA6',
        'gs:uplink:01FA09DF9YWNEX9JG4HEHWFH7P',
        'ns:uplink:01FA09DFA04J4K1MNR9BRR965B',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01FA09DFA0AX0HZF8K14SY5FGD',
        'rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01FA09DFGTZGDEHM5SVR1N8GTH',
      ],
      received_at: '2021-07-07T10:50:00.647244290Z',
      uplink_message: {
        session_key_id: 'AXp0WUMkz75V+jY5OX88DA==',
        f_port: 1,
        f_cnt: 420,
        frm_payload: 'AQ==',
        decoded_payload: {
          bytes: [1],
        },
        rx_metadata: [
          {
            gateway_ids: {
              gateway_id: 'ttig-12e8',
              eui: '58A0CBFFFE8012E8',
            },
            time: '2021-07-07T10:50:00.328195095Z',
            timestamp: 994712508,
            rssi: -89,
            channel_rssi: -89,
            snr: 9,
            uplink_token:
              'ChcKFQoJdHRpZy0xMmU4EghYoMv//oAS6BC8t6jaAxoMCNiNlocGEJy7s7YBIOCs8cv5mQE=',
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
          timestamp: 994712508,
          time: '2021-07-07T10:50:00.328195095Z',
        },
        received_at: '2021-07-07T10:50:00.384736040Z',
        consumed_airtime: '0.046336s',
      },
    },
    correlation_ids: [
      'as:up:01FA09DFHJ0FSWAPWDD3E1RDG6',
      'gs:conn:01FA04C0D25ZM679FXF7QWM6KP',
      'gs:up:host:01FA04C0DNXVBNVWGSE75FVSA6',
      'gs:uplink:01FA09DF9YWNEX9JG4HEHWFH7P',
      'ns:uplink:01FA09DFA04J4K1MNR9BRR965B',
      'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01FA09DFA0AX0HZF8K14SY5FGD',
      'rpc:/ttn.lorawan.v3.NsAs/HandleUplink:01FA09DFGTZGDEHM5SVR1N8GTH',
    ],
    origin: 'ip-redacted.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ', 'RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    unique_id: '01FA09DFK84FJP29A5SGZT4AXP',
  },
  {
    name: 'as.down.data.forward',
    time: '2021-07-07T16:05:48.130282074Z',
    identifiers: [
      {
        device_ids: {
          device_id: 'ttn-uno',
          application_ids: {
            application_id: 'tti-playground',
          },
        },
      },
      {
        device_ids: {
          device_id: 'ttn-uno',
          application_ids: {
            application_id: 'tti-playground',
          },
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ApplicationDownlink',
      f_port: 1,
      frm_payload: 'EjRWeA==',
      correlation_ids: [
        'as:downlink:01FA0VFPY8V67MGSEVGFAZ6TMV',
        'rpc:/ttn.lorawan.v3.AppAs/DownlinkQueuePush:d394eb5c-0d7b-403b-afc1-6da77b84f8c4',
      ],
    },
    correlation_ids: [
      'as:downlink:01FA0VFPY8V67MGSEVGFAZ6TMV',
      'rpc:/ttn.lorawan.v3.AppAs/DownlinkQueuePush:d394eb5c-0d7b-403b-afc1-6da77b84f8c4',
    ],
    origin: 'ip-redacted.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_APPLICATION_TRAFFIC_READ', 'RIGHT_APPLICATION_TRAFFIC_READ'],
    },
    authentication: {
      type: 'Bearer',
      token_type: 'AccessToken',
      token_id: '7IGVW7HUOO6YUHHVB7INWYXFXJSXBXKGINAKYYI',
    },
    remote_ip: '11.111.1.111',
    user_agent:
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36',
    unique_id: '01FA0VFPZ2SZ66BXHT3HZJFAEC',
  },
].reverse()

export const events = [
  ...deviceEvents,
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.create',
    time: '2020-04-23T09:39:04.134034132Z',
    unique_id: '01EK2PTY8Z8X5DNQCE7WDF76H4',
  },
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.delete',
    time: '2020-04-23T09:40:04.134034132Z',
    unique_id: '01EK2PTY278AZ3GYHZ4SZY9BRA',
  },
  {
    identifiers: [{ application_ids: { application_id: 'test-app2' } }],
    name: 'application.update',
    time: '2020-04-23T09:41:04.134034132Z',
    unique_id: '01EK2PTY248S5K98J1JEQ3GYGC',
  },
]

export const gatewayEvents = [
  {
    name: 'gs.down.send',
    time: '2021-04-29T10:14:05.406196093Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.DownlinkMessage',
      raw_payload: 'oGJQCCaAcgUBIGSB4tU=',
      request: {
        downlink_paths: [
          {
            uplink_token:
              'ChsKGQoNdGVrdGVsaWMta29uYRIIZH/a//4Ae9QQ/OH4swwaDAjrjaqEBhCPxrzRAyDgsL3n9a0R',
          },
        ],
        rx1_delay: 5,
        rx1_data_rate_index: 5,
        rx1_frequency: '868100000',
        rx2_data_rate_index: 3,
        rx2_frequency: '869525000',
        frequency_plan_id: 'EU_863_870_TTN',
      },
      correlation_ids: [
        'as:downlink:01F4EHY37VFHS2NH33XWQZ7K8T',
        'gs:conn:01F4C94B13YZRGP62GET7TP2GR',
        'gs:up:host:01F4C94B1BM14V4Y6MER7VA470',
        'gs:uplink:01F4EHY2E8PRMKEQWCB2DV96M5',
        'ns:downlink:01F4EHY3TWB2HBDQ6YBMFV2TZ6',
        'ns:uplink:01F4EHY2F2SQ93K4KPGEHBQT7T',
        'rpc:/ttn.lorawan.v3.AppAs/DownlinkQueueReplace:5f808fba-45f1-44c9-a496-b7d3a516c20e',
        'rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01F4EHY2F2EQ9F0RKDWHSCK5RG',
        'gs:conn:01F4C94B13YZRGP62GET7TP2GR',
        'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01F4EHY3TYKZV59V52SJEMVQMW',
      ],
    },
    correlation_ids: [
      'gs:conn:01F4C94B13YZRGP62GET7TP2GR',
      'rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01F4EHY3TYKZV59V52SJEMVQMW',
    ],
    origin: 'ip-10-23-13-196.eu-west-1.compute.internal',
    context: {
      'tenant-id': '',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01F4EHY3TYN2ESW7P5TKNFQ7T7',
  },
  {
    name: 'gs.up.forward',
    time: '2020-09-25T13:22:00.095877998Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    correlation_ids: [
      'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
      'gs:up:host:01EK0M777XK3QN2Z2ZWV97K5ZZ',
      'gs:uplink:01EK2PTY24Q2WKBB4EHT1NCGPN',
    ],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTY8Z8X5DNQCE7WDF76H4',
  },
  {
    name: 'gs.up.forward',
    time: '2020-09-25T13:21:59.879200087Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    correlation_ids: [
      'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
      'gs:up:host:01EK0M777XJ65ZQXMSQ6GX04N4',
      'gs:uplink:01EK2PTY24Q2WKBB4EHT1NCGPN',
    ],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTY278AZ3GYHZ4SZY9BRA',
  },
  {
    name: 'gs.up.receive',
    time: '2020-09-25T13:21:59.876432885Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QO4oACaAoSBm7vblNX4=',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: '9uU1fg==',
        mac_payload: {
          f_hdr: {
            dev_addr: '260028EE',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 8353,
          },
          f_port: 102,
          frm_payload: '7g==',
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
        frequency: '867700000',
        timestamp: 1136551251,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'tektelic-adrian-ci392',
            eui: '647FDAFFFE0078B8',
          },
          timestamp: 1136551251,
          rssi: -3,
          channel_rssi: -3,
          snr: 8.2,
          uplink_token:
            'CiMKIQoVdGVrdGVsaWMtYWRyaWFuLWNpMzkyEghkf9r//gB4uBDTyvmdBBoMCPfdt/sFEOLv5aEDILiY7/2J8Q8=',
          channel_index: 3,
        },
      ],
      received_at: '2020-09-25T13:21:59.876181474Z',
      correlation_ids: [
        'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
        'gs:uplink:01EK2PTY24Q2WKBB4EHT1NCGPN',
      ],
    },
    correlation_ids: ['gs:conn:01EK0M776NK47SSRM7TTPE27C5', 'gs:uplink:01EK2PTY24Q2WKBB4EHT1NCGPN'],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTY248S5K98J1JEQ3GYGC',
  },
  {
    name: 'gs.up.forward',
    time: '2020-09-25T13:21:51.831451866Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    correlation_ids: [
      'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
      'gs:up:host:01EK0M777XK3QN2Z2ZWV97K5ZZ',
      'gs:uplink:01EK2PTNW4BHATGPAZZTSXX6T8',
    ],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTP6Q30FKAAR9GH0T5XXY',
  },
  {
    name: 'gs.up.forward',
    time: '2020-09-25T13:21:51.496073398Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    correlation_ids: [
      'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
      'gs:up:host:01EK0M777XJ65ZQXMSQ6GX04N4',
      'gs:uplink:01EK2PTNW4BHATGPAZZTSXX6T8',
    ],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTNW8AJ5K9ENV5GGE0NV6',
  },
  {
    name: 'gs.up.receive',
    time: '2020-09-25T13:21:51.492515920Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.UplinkMessage',
      raw_payload: 'QO4oACaCoCADB2ZHoTvPBQ==',
      payload: {
        m_hdr: {
          m_type: 'UNCONFIRMED_UP',
        },
        mic: 'oTvPBQ==',
        mac_payload: {
          f_hdr: {
            dev_addr: '260028EE',
            f_ctrl: {
              adr: true,
            },
            f_cnt: 8352,
            f_opts: 'Awc=',
          },
          f_port: 102,
          frm_payload: 'Rw==',
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
        frequency: '867700000',
        timestamp: 1128122779,
      },
      rx_metadata: [
        {
          gateway_ids: {
            gateway_id: 'tektelic-adrian-ci392',
            eui: '647FDAFFFE0078B8',
          },
          timestamp: 1128122779,
          rssi: -4,
          channel_rssi: -4,
          snr: 9.5,
          uplink_token:
            'CiMKIQoVdGVrdGVsaWMtYWRyaWFuLWNpMzkyEghkf9r//gB4uBCbk/eZBBoMCO/dt/sFEJX43+oBIPiK7srq8A8=',
          channel_index: 3,
        },
      ],
      received_at: '2020-09-25T13:21:51.492305429Z',
      correlation_ids: [
        'gs:conn:01EK0M776NK47SSRM7TTPE27C5',
        'gs:uplink:01EK2PTNW4BHATGPAZZTSXX6T8',
      ],
    },
    correlation_ids: ['gs:conn:01EK0M776NK47SSRM7TTPE27C5', 'gs:uplink:01EK2PTNW4BHATGPAZZTSXX6T8'],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTNW492RXB0MJ2Y0G57P2',
  },
  {
    name: 'gs.status.receive',
    time: '2020-09-25T13:21:47.730010675Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    data: {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.GatewayStatus',
      time: '2020-09-25T13:21:47Z',
      boot_time: '0001-01-01T00:00:00Z',
      versions: {
        'ttn-lw-gateway-server': 'v3.9.4-SNAPSHOT-2d993aba1',
      },
      ip: ['31.201.88.79'],
      metrics: {
        ackr: 20,
        rxfw: 4,
        rxin: 5,
        rxok: 4,
        temp: 60,
        txin: 3,
        txok: 2,
      },
    },
    correlation_ids: ['gs:conn:01EK0M776NK47SSRM7TTPE27C5', 'gs:status:01EK2PTJ6HA7YGQCQ3MZEBZQ92'],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_STATUS_READ'],
    },
    unique_id: '01EK2PTJ6J42DKWGC2F3PZ32TM',
  },
  {
    name: 'gs.down.tx.success',
    time: '2020-09-25T13:21:46.047121750Z',
    identifiers: [
      {
        gateway_ids: {
          gateway_id: 'tektelic-adrian-ci392',
          eui: '647FDAFFFE0078B8',
        },
      },
    ],
    correlation_ids: ['gs:conn:01EK0M776NK47SSRM7TTPE27C5', 'gs:tx_ack:01EK2PTGHZKY8MSVRQPZDB53J2'],
    origin: 'ip-10-20-7-189.eu-west-1.compute.internal',
    context: {
      'tenant-id': 'CgN0dGk=',
    },
    visibility: {
      rights: ['RIGHT_GATEWAY_TRAFFIC_READ'],
    },
    unique_id: '01EK2PTGHZ32F90311PXXGEJQT',
  },
].reverse()

export const organizationEvents = [
  {
    identifiers: [{ organization_ids: { organization_id: 'test-org-id' } }],
    name: 'organization.api-key.create',
    time: '2020-04-23T09:39:04.134034132Z',
    unique_id: '01EK2PTGHZ32F90311PXXGEJQT',
  },
  {
    identifiers: [{ organization_ids: { organization_id: 'test-org-id' } }],
    name: 'organization.update',
    time: '2020-04-23T09:41:04.134034132Z',
    unique_id: '01EK2PTJ6J42DKWGC2F3PZ32TM',
  },
]
