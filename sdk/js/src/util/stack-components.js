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

/* eslint-disable import/prefer-default-export */

import { STACK_COMPONENTS } from './constants'

class StackConfiguration {
  constructor(stackConfig) {
    if (!Boolean(stackConfig)) {
      throw new Error('Stack configuration must be defined')
    }

    const unknownComponents = Object.keys(stackConfig).filter(
      componentName => !STACK_COMPONENTS.includes(componentName),
    )

    if (unknownComponents.length > 0) {
      throw new Error(
        `Cannot instantiate stack configuration with unknown components: ${unknownComponents.join(
          ',',
        )}`,
      )
    }

    this._stackConfig = stackConfig
  }

  /**
   * Selects the url of a stack component.
   * @param {*} componentName - The abbreviation of the component, e.g. is for the Identity Server.
   * @returns {?string} - The url of the component or `undefined` if the component is not available.
   */
  getComponentUrlByName(componentName) {
    return this._stackConfig[componentName]
  }

  /**
   * Selects the hostname of a stack component.
   * @param {*} componentName - The abbreviation of the component, e.g. is for the Identity Server.
   * @returns {?string} - The hostname of the component address or `undefined` if the component is not available.
   */
  getComponentHostByName(componentName) {
    try {
      const url = this.getComponentUrlByName(componentName)

      return new URL(url).hostname
    } catch (error) {
      // do not propagate the error, simply return `undefined`
    }
  }

  /**
   * Checks whether a stack component is available in the configuration.
   * @param {*} componentName - The abbreviation of the component, e.g. is for the Identity Server.
   * @returns {boolean} - `true` if the component is available in the configuration, `false` otherwise.
   */
  isComponentAvailable(componentName) {
    const componentUrl = this.getComponentUrlByName(componentName)

    return typeof componentUrl === 'string' && componentUrl.length > 0
  }

  /**
   * Identity Server hostname getter.
   * @returns {?string} - The hostname of the Identity Server of the stack configuration.
   */
  get isHost() {
    return this.isComponentAvailable('is') && this.getComponentHostByName('is')
  }

  /**
   * Network Server hostname getter.
   * @returns {?string} - The hostname of the Network Server of the stack configuration.
   */
  get nsHost() {
    return this.isComponentAvailable('ns') && this.getComponentHostName('ns')
  }

  /**
   * Application Server hostname getter.
   * @returns {?string} - The hostname of the Application Server of the stack configuration.
   */
  get asHost() {
    return this.isComponentAvailable('as') && this.getComponentHostName('as')
  }

  /**
   * Join Server hostname getter.
   * @returns {?string} - The hostname of the Join Server of the stack configuration.
   */
  get jsHost() {
    return this.isComponentAvailable('js') && this.getComponentHostByName('js')
  }

  /**
   * Avaible stack components getter.
   * @returns {Array<string>} - A list of avaiable component abbreviations, e.g. [is, as, ns, js].
   */
  get availableComponents() {
    return Object.keys(this._stackConfig)
  }

  /** Takes a list of allowed components and only returns components that have
   * distinct base urls. Used to subscribe to event streaming sources when the
   * stack uses multiple hosts.
   * @param {Array<string>} components - A list of abbreviations of stack components to return distinct ones from.
   * @returns {Array<string>} - An array of components that have distinct base urls.
   */
  getComponentsWithDistinctBaseUrls(components = STACK_COMPONENTS) {
    const distinctComponents = components.reduce((collection, component) => {
      if (
        Boolean(this._stackConfig.isComponentAvailable(component)) &&
        !Object.values(collection).includes(this._stackConfig.getComponentUrlByName(component))
      ) {
        return { ...collection, [component]: this._stackConfig.getComponentUrlByName(component) }
      }
      return collection
    }, {})

    return Object.keys(distinctComponents)
  }
}

export default StackConfiguration
