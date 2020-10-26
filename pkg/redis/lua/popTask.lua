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

-- ARGV[1] - group
-- ARGV[2] - consumer
-- ARGV[3] - pivot
-- ARGV[4] - timeout
--
-- KEYS[1] - ready task key
-- KEYS[2] - input task key
-- KEYS[3] - waiting task key

local function format_ready(xs)
  local ret = { 'ready', 'id', xs[1][2][1][1] }
  for i, v in ipairs(xs[1][2][1][2]) do
    ret[i+3] = v
  end
  return ret
end

local xs = redis.call('xreadgroup', 'group', ARGV[1], ARGV[2], 'count', '1', 'streams', KEYS[1], '>')
if xs then
  return format_ready(xs)
end

xs = redis.call('xreadgroup', 'group', ARGV[1], ARGV[2], 'NOACK', 'streams', KEYS[2], '>')
if xs then
  local zs = {}
  for i, x in ipairs(xs[1][2]) do
    local start_at, payload
    for j=1,#x[2],2 do
      local name = x[2][j]
      if     name == 'start_at' then start_at = x[2][j+1]
      elseif name == 'payload'  then payload  = x[2][j+1]
      end
    end
    zs[#zs+1] = start_at
    zs[#zs+1] = payload
  end
  redis.call('zadd', KEYS[3], unpack(zs))
end

local zs = redis.call('zrangebyscore', KEYS[3], '-inf', ARGV[3], 'withscores')
if #zs > 0 then
  local members = {}
  for i=1,#zs,2 do
    local member = zs[i]
    members[#members+1] = member
    redis.call('xadd', KEYS[1], '*', 'payload', member, 'start_at', zs[i+1])
  end
  redis.call('zrem', KEYS[3], unpack(members))
  return format_ready(redis.call('xreadgroup', 'group', ARGV[1], ARGV[2], 'count', '1', 'streams', KEYS[1], '>'))
end

local ret = { 'waiting' }
zs = redis.call('zrangebyscore', KEYS[3], '-inf', '+inf', 'withscores', 'limit', '0', '1')
if #zs > 0 then
  ret[#ret+1] = 'next_at'
  ret[#ret+1] = zs[2]
end

xs = redis.call('xrevrange', KEYS[2], '+', '-', 'count', '1')
if xs then
  ret[#ret+1] = 'last_id'
  ret[#ret+1] = xs[1][1]
end
return ret
