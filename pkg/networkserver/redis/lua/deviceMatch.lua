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

-- ARGV[1]	- 2 LSB of FCnt (same as 16-bit FCnt field in MAC frames)
-- ARGV[2] 	- output TTL in milliseconds
--
-- KEYS[1] 	- previous matching result key
--
-- KEYS[2] 	- sorted set of uids of devices matching current session DevAddr sorted ascending by LSB of LastFCntUp
-- KEYS[3] 	- hash containing msgpack-encoded sessions for devices matching current session DevAddr keyed by uid
-- KEYS[4] 	- sorted list of uids of devices matching with current session LastFCntUp LSB being lower than or equal to uplink LSB
-- KEYS[5] 	- sorted list of uids of devices matching with current session LastFCntUp LSB being greater than uplink LSB
-- KEYS[6]  - copy of KEYS[3]
--
-- KEYS[7] 	- sorted set of uids of devices matching pending session DevAddr sorted ascending by creation time
-- KEYS[8] 	- hash containing msgpack-encoded sessions for devices matching pending session DevAddr keyed by uid
-- KEYS[9]  - sorted list of uids of devices matching pending session DevAddr
-- KEYS[10] - copy of KEYS[8]
if redis.call('pexpire', KEYS[1], ARGV[2]) == 1 then
  return { 'result', redis.call('get', KEYS[1]) }
end

-- Update expiration of matching keys - if any exist - return.
local to_scan = { 'scan' }
local function scan_expiring(i)
  local ret = redis.call('pexpire', KEYS[i], ARGV[2])
  if ret == 1 then
    table.insert(to_scan, i)
  end
  return ret
end
if scan_expiring(4) + scan_expiring(5) > 0 then
  redis.call('pexpire', KEYS[6], ARGV[2])
end
if #KEYS == 10 then
  if scan_expiring(9) > 0 then
    redis.call('pexpire', KEYS[10], ARGV[2])
  end
end
if #to_scan > 1 then
  return to_scan
end

local pivot = redis.call('zcount', KEYS[2], '-inf', ARGV[1])
if pivot > 0 then
  redis.call('sort', KEYS[2], 'by', 'nosort', 'limit', 0, pivot, 'store', KEYS[4])
  redis.call('pexpire', KEYS[4], ARGV[2])
  table.insert(to_scan, 4)
end
local gt = redis.call('sort', KEYS[2], 'by', 'nosort', 'limit', pivot, -1, 'store', KEYS[5])
if gt > 0 then
  redis.call('pexpire', KEYS[5], ARGV[2])
  table.insert(to_scan, 5)
end
if pivot > 0 or gt > 0 then
  -- TODO: Use COPY once Redis is updated to 6.2.0 (https://github.com/TheThingsNetwork/lorawan-stack/issues/3592)
  -- redis.call('copy', KEYS[3], KEYS[6])
  -- redis.call('pexpire', KEYS[6], ARGV[2])
  redis.call('restore', KEYS[6], ARGV[2], redis.call('dump', KEYS[3]))
end

if #KEYS == 10 and redis.call('sort', KEYS[7], 'by', 'nosort', 'store', KEYS[9]) > 0 then
  redis.call('pexpire', KEYS[9], ARGV[2])
  table.insert(to_scan, 9)
  -- TODO: Use COPY once Redis is updated to 6.2.0 (https://github.com/TheThingsNetwork/lorawan-stack/issues/3592)
  -- redis.call('copy', KEYS[8], KEYS[10])
  -- redis.call('pexpire', KEYS[10], ARGV[2])
  redis.call('restore', KEYS[10], ARGV[2], redis.call('dump', KEYS[8]))
end

if #to_scan > 1 then
    return to_scan
end
return nil
