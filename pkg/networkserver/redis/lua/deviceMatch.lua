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
-- KEYS[6] 	- sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 16-bit frame counters
-- KEYS[7] 	- sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 16-bit frame counters being processed
--
-- KEYS[8] 	- sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 32-bit frame counters with no rollover
-- KEYS[9] 	- sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 32-bit frame counters with no rollover being processed
--
-- KEYS[10]	- sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 32-bit frame counters with rollover
-- KEYS[11] - sorted list of uid, LastFCntUp, FNwkSIntKey, MACVersion of matching devices using 32-bit frame counters with rollover being processed
--
-- KEYS[12] - list of uids of devices matching pending session DevAddr
-- KEYS[13] - list of uids of devices matching pending session DevAddr being processed
--
-- KEYS[14] - list of uids of devices matching either current or pending session DevAddr not present in either KEYS[2], KEYS[3], nor KEYS[4]
-- KEYS[15] - list of uids of devices matching either current or pending session DevAddr not present in either KEYS[2], KEYS[3], nor KEYS[4] being processed
-- NOTE: It is expected that count of devices using 16-bit frame counters << count of devices using 32-bit frame counters.
if redis.call('pexpire', KEYS[1], ARGV[2]) > 0 then
  return redis.call('get', KEYS[1])
end

-- Update expiration of all match keys - if any exist - return.
-- `+` operator is used to avoid potential short-circuit evaluation.
if redis.call('pexpire', KEYS[6], ARGV[2]) +
   redis.call('pexpire', KEYS[7], ARGV[2]) +
   redis.call('pexpire', KEYS[8], ARGV[2]) +
   redis.call('pexpire', KEYS[9], ARGV[2]) +
   redis.call('pexpire', KEYS[10], ARGV[2]) +
   redis.call('pexpire', KEYS[11], ARGV[2]) +
   redis.call('pexpire', KEYS[12], ARGV[2]) +
   redis.call('pexpire', KEYS[13], ARGV[2]) +
   redis.call('pexpire', KEYS[14], ARGV[2]) +
   redis.call('pexpire', KEYS[15], ARGV[2]) > 0 then
	return nil
end

local n, sorted = 0, 0
local fromCurrent = '('..ARGV[1]

local shortIdx = redis.call('zcount', KEYS[2], fromCurrent, '+inf')
sorted = redis.call('sort', KEYS[2], 'by', 'nosort', 'limit', 0, shortIdx, 'store', KEYS[6])
if sorted > 0 then
  n += sorted
  redis.call('pexpire', KEYS[6], ARGV[2])
end

local longIdx = redis.call('zcount', KEYS[3], fromCurrent, '+inf')
sorted = redis.call('sort', KEYS[3], 'by', 'nosort', 'limit', 0, longIdx, 'store', KEYS[8])
if sorted > 0 then
  n += sorted
  redis.call('pexpire', KEYS[8], ARGV[2])
end

sorted = redis.call('sort', KEYS[3], 'by', 'nosort', 'limit', longIdx, -1, 'store', KEYS[10])
if sorted > 0 then
  n += sorted
  redis.call('pexpire', KEYS[10], ARGV[2])
end

sorted = redis.call('sort', KEYS[4], 'by', 'nosort', 'store', KEYS[12])
if sorted > 0 then
  n += sorted
  redis.call('pexpire', KEYS[12], ARGV[2])
end

sorted = redis.call('sort', KEYS[5], 'by', 'nosort', 'store', KEYS[14])
if sorted > 0 then
  n += sorted
  redis.call('pexpire', KEYS[14], ARGV[2])
end

return sorted
