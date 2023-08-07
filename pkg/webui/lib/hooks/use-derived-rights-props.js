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

import { useMemo } from 'react'

import { RIGHT_ALL } from '@console/lib/rights'

/**
 * `useComputedProps` is a hook that can be used to pass-in props that are
 * derived from the other props of the components. This is useful to ensure that
 * expensive prop computations only need to be done once upon prop changes.
 *
 * @param {object} props - The props that are used to compute the computed props.
 * @returns {object} - The computed props.
 */
const useDerivedRightProps = props => {
  const { value, pseudoRight: grantablePseudoRight, rights: grantableRights } = props

  return useMemo(() => {
    // Extract the pseudo right from own rights or granted rights.
    let derivedPseudoRight = []
    if (grantablePseudoRight && !Array.isArray(grantablePseudoRight)) {
      derivedPseudoRight = [grantablePseudoRight]
    } else if (grantablePseudoRight && Array.isArray(grantablePseudoRight)) {
      derivedPseudoRight = grantablePseudoRight
    } else {
      derivedPseudoRight = value.filter(right => right !== RIGHT_ALL && right.endsWith('_ALL'))
    }
    // Filter out rights that the entity has but may not be granted by the user.
    const outOfOwnScopeRights = !Boolean(grantablePseudoRight)
      ? value.filter(right => !grantableRights.includes(right))
      : []

    // Extract all rights by combining granted and grantable rights.
    const derivedRights = [...grantableRights, ...outOfOwnScopeRights].sort()

    // Store whether out of scope pseudo rights are present.
    const hasOutOfOwnScopePseudoRight =
      outOfOwnScopeRights.filter(right => right.endsWith('_ALL')).length !== 0

    // Store granted individual rights.
    const grantedIndividualRights = value.filter(right => !derivedPseudoRight.includes(right))

    // Store out of own scope individual rights.
    const outOfOwnScopeIndividualRights = !Boolean(grantablePseudoRight)
      ? grantedIndividualRights.filter(right => !grantableRights.includes(right))
      : []

    // Determine whether a pseudo right is granted.
    const hasPseudoRightGranted =
      value.includes(RIGHT_ALL) ||
      derivedPseudoRight.some(derivedRight => value.includes(derivedRight))

    // Determine the current grant type.
    const grantType = hasPseudoRightGranted ? 'pseudo' : 'individual'

    return {
      outOfOwnScopeIndividualRights,
      hasOutOfOwnScopePseudoRight,
      derivedPseudoRight,
      derivedRights,
      hasPseudoRightGranted,
      grantType,
    }
  }, [value, grantablePseudoRight, grantableRights])
}

export default useDerivedRightProps
