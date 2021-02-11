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

import Applications from './service/applications'

import TTS from '.'

const mockApplicationData = {
  ids: {
    application_id: 'test',
  },
  created_at: '2018-08-29T14:00:20.793Z',
  updated_at: '2018-08-29T14:00:20.793Z',
  name: 'string',
  description: 'string',
  attributes: {
    additionalProp1: 'string',
    additionalProp2: 'string',
    additionalProp3: 'string',
  },
  contact_info: [
    {
      contact_type: 'CONTACT_TYPE_OTHER',
      contact_method: 'CONTACT_METHOD_OTHER',
      value: 'string',
      public: true,
      validated_at: '2018-08-29T14:00:20.793Z',
    },
  ],
}

const mockDeviceData = {
  ids: {
    device_id: 'test-device',
    application_ids: {
      application_id: 'test',
    },
    dev_eui: 'string',
    join_eui: 'string',
    dev_addr: 'string',
  },
}

jest.mock('./api', function () {
  return jest.fn().mockImplementation(function () {
    return {
      ApplicationRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockApplicationData }),
        List: jest.fn().mockResolvedValue({
          data: { applications: [mockApplicationData] },
          headers: { 'x-total-count': 1 },
        }),
      },
      EndDeviceRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockDeviceData }),
      },
      NsEndDeviceRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockDeviceData }),
      },
      AsEndDeviceRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockDeviceData }),
      },
      JsEndDeviceRegistry: {
        Get: jest.fn().mockResolvedValue({ data: mockDeviceData }),
      },
    }
  })
})

describe('SDK class', function () {
  const token = 'faketoken'
  const tts = new TTS({
    authorization: {
      mode: 'key',
      key: token,
    },
    connectionType: 'http',
    stackConfig: { is: 'http://localhost:1885/api/v3' },
  })

  it('instanciates successfully', async function () {
    expect(tts).toBeDefined()
    expect(tts).toBeInstanceOf(TTS)
    expect(tts.Applications).toBeInstanceOf(Applications)
  })

  it('retrieves application instance correctly', async function () {
    const app = await tts.Applications.getById('test')
    expect(app).toBeDefined()
  })
})
