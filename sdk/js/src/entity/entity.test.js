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

import Entity from './entity'

const mockData = {
  a: {
    b: {
      c: 'test',
    },
    d: 'test',
  },
  e: 'test',
  f: 'left-unchanged',
}

describe('Entity', function() {
  test('recursively proxies ingested data', function() {
    const entity = new Entity(mockData)

    expect(entity).toBeInstanceOf(Entity)
    expect(entity.toObject()).toMatchObject(mockData)
    expect(entity.a._changed).toHaveLength(0)
  })

  test('keeps track of changes', function() {
    const entity = new Entity(mockData)

    entity.e = 'foo'
    entity.a.b.c = 'bar'
    entity.a.b.c = 'again'
    entity.a.d = 'baz'

    expect(entity._changed).toEqual(['e'])
    expect(entity.a._changed).toEqual(['d'])
    expect(entity.a.b._changed).toEqual(['c'])
    expect(entity._rawData).toMatchObject(mockData)

    expect(entity.getUpdateMask()).toEqual(['a.b.c', 'a.d', 'e'])
  })

  test('clears changes correctly', function() {
    const entity = new Entity(mockData)

    entity.e = 'foo'
    expect(entity.getUpdateMask()).toEqual(['e'])
    entity.clearValues()
    expect(entity.getUpdateMask()).toEqual([])
  })
})
