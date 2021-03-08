-- Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

-- The following script implements XAUTOCLAIM in Lua. See https://redis.io/commands/xautoclaim for documentation.
-- JUSTID is not supported and COUNT is mandatory.
--
-- KEYS[1] - key
-- ARGV[1] - group
-- ARGV[2] - consumer
-- ARGV[3] - min-idle-time
-- ARGV[4] - start
-- ARGV[5] - count

local xps = redis.call('xpending', KEYS[1], ARGV[1], ARGV[4], '+', ARGV[5])
if not xps then
  return nil
end

local ids = {}
for _, xp in ipairs(xps) do
  if xp[3] >= tonumber(ARGV[3]) then
    ids[#ids+1] = xp[1]
  end
end
if #ids == 0 then
  return nil
end

return {
  ids[#ids],
  redis.call('xclaim', KEYS[1], ARGV[1], ARGV[2], ARGV[3], unpack(ids)),
}
