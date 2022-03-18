-- Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
--
-- The following script drains KEYS[2] and adds read entries to KEYS[3].
-- If, in result, there are tasks in KEYS[3] with timestamp less than or equal to ARGV[3], it moves all those tasks to KEYS[1].
-- It returns the dispatch time of the earliest undispatched task, if such exist.
--
-- ARGV[1] - group ID
-- ARGV[2] - consumer ID
-- ARGV[3] - pivot - current time, expressed as nanoseconds elapsed since Unix epoch
-- ARGV[4] - approximate maximum length of ready task stream
--
-- KEYS[1] - ready task key
-- KEYS[2] - input task key
-- KEYS[3] - waiting task key
--
-- The "unpack" Lua function may not unpack more elements than the max stack length.
-- In order to avoid this natural limitation, we will periodically flush the waiting keys
-- as they are moved into the ready stream.
local max_unpack = 512

-- Claim any entries which have not been processed yet.
redis.call('xautoclaim', KEYS[2], ARGV[1], ARGV[2], '0', '0-0', 'justid')

-- Drain the input task stream (KEYS[2]) into the waiting task sorted set (KEYS[3]).
-- We drain both our pending entries (using the 0-0 message ID), and any newer entries (using the > message ID).
local streams = redis.call('xreadgroup', 'group', ARGV[1], ARGV[2], 'noack', 'streams', KEYS[2], KEYS[2], '0-0', '>')
if #streams > 0 then
    for i, xs in ipairs(streams) do
        local messages = xs[2]
        for j, x in ipairs(messages) do
            -- We need to explicitly check if the fields of the message exist, since XAUTOCLAIM
            -- may have claimed deleted messages.
            -- TODO: Starting with Redis 7.0, XAUTOCLAIM will automatically skip these messages.
            -- Remove the nil check (https://github.com/TheThingsNetwork/lorawan-stack/issues/5269).
            local fields = x[2]
            if fields ~= nil then
                local start_at, payload, replace
                for k = 1, #fields, 2 do
                    local name = fields[k]
                    if name == 'start_at' then
                        start_at = fields[k + 1]
                    elseif name == 'payload' then
                        payload = fields[k + 1]
                    elseif name == 'replace' then
                        replace = fields[k + 1]
                    end
                end
                if replace then
                    redis.call('zadd', KEYS[3], start_at, payload)
                else
                    redis.call('zadd', KEYS[3], 'nx', start_at, payload)
                end
            end

            -- NOACK affects only messages which are not already in the pending entries list.
            -- As such, we need to manually acknowledge these messages.
            if i == 1 then
                redis.call('xack', KEYS[2], ARGV[1], x[1])
            end
        end
    end
end

-- Find the tasks whose score is smaller or equal to the pivot, which now must be dispatched (moved to the ready stream, KEYS[1]).
local zs = redis.call('zrange', KEYS[3], '-inf', ARGV[3], 'withscores', 'byscore')
if #zs > 0 then
    local members = {}
    for i = 1, #zs, 2 do
        if #members > max_unpack then
            redis.call('zrem', KEYS[3], unpack(members))
            members = {}
        end

        local member = zs[i]
        members[#members + 1] = member
        redis.call('xadd', KEYS[1], 'maxlen', '~', ARGV[4], '*', 'payload', member, 'start_at', zs[i + 1])
    end
    redis.call('zrem', KEYS[3], unpack(members))
end

-- Find the earliest task which may be dispatched in the future.
-- The caller can then block waiting for a new task in the input task stream (KEYS[2])
-- or for the time until the next task to pass.
zs = redis.call('zrange', KEYS[3], '-inf', '+inf', 'withscores', 'byscore', 'limit', 0, 1)
if #zs > 0 then
    return zs[2]
end
