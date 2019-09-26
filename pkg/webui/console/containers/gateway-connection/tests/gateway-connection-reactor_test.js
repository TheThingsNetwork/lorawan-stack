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

import getConnectionReactorDriver from './gateway-connection-reactor_driver'

const createEvent = (name, time, identifiers) => ({
  name,
  time,
  identifiers,
})

const getEventFixture = () => createEvent('test.name', '2019-09-10T07:30:14.232137918Z', ['id'])

describe('GatewayConnectionReactor', () => {
  let driver = null
  const baseProps = {
    updateGatewayStatistics: () => {},
  }

  beforeEach(() => {
    driver = getConnectionReactorDriver()
  })

  it('should match snapshot', () => {
    const props = {
      ...baseProps,
      latestEvent: getEventFixture(),
      componentProp: 'prop',
    }

    driver.when.created(props)
    expect(driver.component).toMatchSnapshot()
  })

  it('should render wrapped component', () => {
    driver.when.created(baseProps)
    expect(driver.is.wrappedComponentPresent()).toBe(true)
  })

  it('should pass wrapped component props', () => {
    const props = {
      ...baseProps,
      componentProp: 'prop',
    }

    driver.when.created(props)
    const wrappedComponentProps = driver.get.wrappedComponentProps()
    expect(wrappedComponentProps.updateGatewayStatistics).toBeUndefined()
    expect(wrappedComponentProps.componentProp).toBe(props.componentProp)
  })

  it('should call `updateGatewayStatistics` on new event of correct type', () => {
    const latestEvent = getEventFixture()
    const updateGatewayStatistics = jest.fn()
    const props = {
      latestEvent,
      updateGatewayStatistics,
    }

    driver.when.created(props)
    expect(driver.is.wrappedComponentPresent()).toBe(true)
    expect(updateGatewayStatistics.mock.calls).toHaveLength(0)

    const newLatestEvent = getEventFixture()
    let newProps = {
      latestEvent: newLatestEvent,
      updateGatewayStatistics,
    }

    driver.when.updated(newProps)
    expect(updateGatewayStatistics.mock.calls).toHaveLength(0)

    newProps = {
      updateGatewayStatistics,
      latestEvent: {
        ...getEventFixture(),
        name: createEvent('gs.up.receive').name,
      },
    }

    driver.when.updated(newProps)
    expect(updateGatewayStatistics.mock.calls).toHaveLength(1)

    newProps = {
      updateGatewayStatistics,
      latestEvent: {
        ...getEventFixture(),
        name: createEvent('gs.down.send').name,
      },
    }

    driver.when.updated(newProps)
    expect(updateGatewayStatistics.mock.calls).toHaveLength(2)

    newProps = {
      updateGatewayStatistics,
      latestEvent: {
        ...getEventFixture(),
        name: createEvent('gs.status.receive').name,
      },
    }

    driver.when.updated(newProps)
    expect(updateGatewayStatistics.mock.calls).toHaveLength(3)
  })
})
