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

import { hasSpecial, hasUpper, hasDigit, hasMinLength, hasMaxLength } from '.'

describe('Password utilities', () => {
  describe('hasSpecial', () => {
    const tests = [
      { result: true },
      { password: '$', special: 1, result: true },
      { password: '𝒬', special: 1, result: true },
      { password: 'test-password', special: 1, result: true },
      { password: 'test-password', special: 2, result: false },
      { password: 'test-p@ssword', special: 2, result: true },
      { password: 'testpassword1', special: 1, result: false },
      { password: 'Testpassword1', special: 1, result: false },
      { password: 'testpǟssword', special: 1, result: false },
    ]

    it.each(tests.map(({ password, special, result }) => [password, special, result]))(
      'yields hasSpecial(%p, %d) = %p',
      (pw, special, expected) => {
        expect(hasSpecial(pw, special)).toBe(expected)
      },
    )
  })

  describe('hasUpper', () => {
    const tests = [
      { result: true },
      { password: 'no-upper', upper: 0, result: true },
      { password: 'A', upper: 1, result: true },
      { password: 'Й', upper: 1, result: true },
      { password: 'Θ', upper: 1, result: true },
      { password: '𝔉', upper: 1, result: true },
      { password: '𝒬𝔉Z', upper: 3, result: true },
      { password: 'θ', upper: 1, result: false },
      { password: '1', upper: 1, result: false },
      { password: 'test-password', upper: 1, result: false },
    ]

    it.each(tests.map(({ password, upper, result }) => [password, upper, result]))(
      'yields hasUpper(%p, %d) = %p',
      (pw, upper, expected) => {
        expect(hasUpper(pw, upper)).toBe(expected)
      },
    )
  })

  describe('hasDigit', () => {
    const tests = [
      { result: true },
      { password: 'no-digits', digits: 0, result: true },
      { password: '1', digits: 0, result: true },
      { password: '1', digits: 1, result: true },
      { password: '٩', digits: 1, result: true },
      { password: '1️⃣', digits: 1, result: true },
      { password: 'test-password٩', digits: 1, result: true },
      { password: 'test-password1٩', digits: 2, result: true },
      { password: 'no-digits', digits: 1, result: false },
      { password: '1', digits: 2, result: false },
      { password: '٩', digits: 2, result: false },
      { password: '1️⃣', digits: 2, result: false },
    ]

    it.each(tests.map(({ password, digits, result }) => [password, digits, result]))(
      'yields hasDigit(%p, %d) = %p',
      (pw, digits, expected) => {
        expect(hasDigit(pw, digits)).toBe(expected)
      },
    )
  })

  describe('hasMinLength', () => {
    const testPassword = 'test-password'
    const testPassword2 = 'password🤫'

    const tests = [
      { result: true },
      { password: testPassword, minLength: 0, result: true },
      { password: testPassword, minLength: testPassword.length, result: true },
      { password: testPassword, minLength: testPassword.length - 1, result: true },
      { password: testPassword, minLength: testPassword.length + 1, result: false },
      { password: testPassword2, minLength: testPassword2.length, result: true },
      { password: testPassword2, minLength: testPassword2.length - 1, result: true },
      { password: testPassword2, minLength: testPassword2.length + 1, result: false },
    ]

    it.each(tests.map(({ password, minLength, result }) => [password, minLength, result]))(
      'yields hasMaxLength(%p, %d) = %p',
      (pw, minLength, expected) => {
        expect(hasMinLength(pw, minLength)).toBe(expected)
      },
    )
  })

  describe('hasMaxLength', () => {
    const testPassword = 'test-password'
    const testPassword2 = 'password🤫'

    const tests = [
      { result: true },
      { password: 'a'.repeat(100), result: true },
      { password: testPassword, maxLength: testPassword.length, result: true },
      { password: testPassword, maxLength: testPassword.length + 1, result: true },
      { password: testPassword, maxLength: testPassword.length - 1, result: false },
      { password: testPassword2, maxLength: testPassword2.length, result: true },
      { password: testPassword2, maxLength: testPassword2.length + 1, result: true },
      { password: testPassword2, maxLength: testPassword2.length - 1, result: false },
    ]

    it.each(tests.map(({ password, maxLength, result }) => [password, maxLength, result]))(
      'yields hasMaxLength(%p, %d) = %p',
      (pw, maxLength, expected) => {
        expect(hasMaxLength(pw, maxLength)).toBe(expected)
      },
    )
  })
})
