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

if #ARGV == 2 then
	redis.call("lrem", KEYS[2], 1, ARGV[2])
end


for i = 1, #KEYS do
  local uid
  if KEYS[i]:sub(-10) == "processing" then
	  uid = redis.call('rpop', KEYS[i])
  else
	  uid = redis.call('rpoplpush', KEYS[i], KEYS[i+1])
  end
	if uid then
	  for j = i, #KEYS, 1 do
	  	redis.call('pexpire', KEYS[j], ARGV[1])
	  end
	  return {i,uid}
	end
end
return nil
