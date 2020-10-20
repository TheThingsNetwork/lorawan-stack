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
-- KEYS[2] 	- sorted set of uids of devices matching current session DevAddr using 16-bit frame counters sorted by LastFCntUp
-- KEYS[3] 	- sorted set of uids of devices matching current session DevAddr using 32-bit frame counters sorted by 2 LSB of LastFCntUp
-- KEYS[4] 	- set of uids of devices matching pending session DevAddr
-- KEYS[5]  - set of uids of devices matching either current or pending session DevAddr (legacy)
--
-- KEYS[6] 	- sorted list of uids of devices matching using 16-bit frame counters
-- KEYS[7] 	- sorted list of uids of devices matching using 16-bit frame counters being processed
--
-- KEYS[8] 	- sorted list of uids of devices matching using 32-bit frame counters with no rollover
-- KEYS[9] 	- sorted list of uids of devices matching using 32-bit frame counters with no rollover being processed
--
-- KEYS[10] - list of uids of devices matching pending session DevAddr
-- KEYS[11] - list of uids of devices matching pending session DevAddr being processed
--
-- KEYS[12]	- sorted list of uids of devices matching using 32-bit frame counters with rollover
-- KEYS[13] - sorted list of uids of devices matching using 32-bit frame counters with rollover being processed
--
-- KEYS[14] - sorted list of uids of devices matching using 16-bit frame counters with a reset
-- KEYS[15] - sorted list of uids of devices matching using 16-bit frame counters with a reset being processed
--
-- KEYS[16] - list of uids of devices matching either current or pending session DevAddr not present in either KEYS[2], KEYS[3], nor KEYS[4]
-- KEYS[17] - list of uids of devices matching either current or pending session DevAddr not present in either KEYS[2], KEYS[3], nor KEYS[4] being processed
-- NOTE: The script is optimized for the assumption that count of devices using 16-bit frame counters << count of devices using 32-bit frame counters.
if redis.call('pexpire', KEYS[1], ARGV[2]) > 0 then
  return redis.call('get', KEYS[1])
end

-- Update expiration of all match keys - if any exist - return.
local toScan = {}
for i=6,17 do
  if redis.call('pexpire', KEYS[i], ARGV[2]) == 1 then
    table.insert(toScan, i)
  end
end
if #toScan > 0 then
    return toScan
end

local shortCount = redis.call('zcount', KEYS[2], '-inf', ARGV[1])
if redis.call('sort', KEYS[2], 'by', 'nosort', 'limit', 0, shortCount, 'store', KEYS[6]) > 0 then
  redis.call('pexpire', KEYS[6], ARGV[2])
  table.insert(toScan, 6)
end

local longCount = redis.call('zcount', KEYS[3], '-inf', ARGV[1])
if redis.call('sort', KEYS[3], 'by', 'nosort', 'limit', 0, longCount, 'store', KEYS[8]) > 0 then
  redis.call('pexpire', KEYS[8], ARGV[2])
  table.insert(toScan, 8)
end

if redis.call('sort', KEYS[4], 'by', 'nosort', 'store', KEYS[10]) > 0 then
  redis.call('pexpire', KEYS[10], ARGV[2])
  table.insert(toScan, 10)
end

if redis.call('sort', KEYS[3], 'by', 'nosort', 'limit', longCount, -1, 'store', KEYS[12]) > 0 then
  redis.call('pexpire', KEYS[12], ARGV[2])
  table.insert(toScan, 12)
end

if redis.call('sort', KEYS[2], 'by', 'nosort', 'limit', shortCount, -1, 'store', KEYS[14]) > 0 then
  redis.call('pexpire', KEYS[14], ARGV[2])
  table.insert(toScan, 14)
end

if redis.call('sort', KEYS[5], 'by', 'nosort', 'store', KEYS[16]) > 0 then
  redis.call('pexpire', KEYS[16], ARGV[2])
  table.insert(toScan, 16)
end

if #toScan > 0 then
    return toScan
end
return nil
