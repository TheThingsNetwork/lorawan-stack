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

local ack = ARGV[1]
table.remove(ARGV, 1)
for _, old_uid in ipairs(ARGV) do
  local uid = redis.call('lindex', KEYS[1], -1)
  if not uid then
    return nil
  end
  if uid ~= old_uid then
    local s = redis.call('hget', KEYS[2], uid)
    local m = cmsgpack.unpack(s)
    if not m.supports_32_bit_f_cnt or m.supports_32_bit_f_cnt.value
      or ack == 0 and m.resets_f_cnt and m.resets_f_cnt.value then
      return { uid, s }
    end
  end
  redis.call('ltrim', KEYS[1], 0, -2)
  redis.call('hdel', KEYS[2], uid)
end

local uid = redis.call('lindex', KEYS[1], -1)
while uid do
  local s = redis.call('hget', KEYS[2], uid)
  local m = cmsgpack.unpack(s)
  if not m.supports_32_bit_f_cnt or m.supports_32_bit_f_cnt.value
    or ack == 0 and m.resets_f_cnt and m.resets_f_cnt.value then
    return { uid, s }
  end
  redis.call('ltrim', KEYS[1], 0, -2)
  redis.call('hdel', KEYS[2], uid)
  uid = redis.call('lindex', KEYS[1], -1)
end
return nil
