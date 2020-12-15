-- Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

for old_uid in ARGV do
  local uid = redis.call('lindex', KEYS[1], 0)
  if uid ~= old_uid then
    return uid
  end
  redis.call('ltrim', KEYS[1], 1, -1)
  redis.call('hdel', KEYS[2], ARGV[1])
end
return redis.call('lindex', KEYS[1], 0)
