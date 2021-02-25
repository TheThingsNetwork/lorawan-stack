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

export const isValidProfilePictureObject = profilePicture =>
  Boolean(profilePicture) && Boolean(profilePicture.sizes)

export const getProfilePictureSizes = profilePicture => {
  if (!isValidProfilePictureObject(profilePicture)) {
    return {}
  }

  return profilePicture.sizes
}

export const getSmallestAvailableProfilePicture = profilePicture =>
  getClosestProfilePictureBySize(profilePicture, 64)

export const getOriginalSizeProfilePicture = profilePicture => {
  const sizes = getProfilePictureSizes(profilePicture)

  return sizes[0]
}

export const getClosestProfilePictureBySize = (profilePicture, size) => {
  const sizes = getProfilePictureSizes(profilePicture)
  if (sizes[size]) {
    return sizes[size]
  }

  const closestSize = Object.keys(sizes).sort((a, b) => Math.abs(size - a) - Math.abs(size - b))[0]

  return sizes[closestSize]
}

export const isGravatarProfilePicture = profilePicture => {
  if (!isValidProfilePictureObject(profilePicture)) return false
  const sizes = profilePicture.sizes

  return sizes[64] ? sizes[64].startsWith('https://www.gravatar.com') : false
}

export const convertUriToProfilePicture = uri => ({ sizes: { 0: uri } })
