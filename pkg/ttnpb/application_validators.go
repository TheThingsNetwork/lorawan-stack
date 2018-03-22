// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// Validate is used as validator function by the GRPC validator interceptor.
func (req *DownlinkQueueRequest) Validate() error {
	return req.EndDeviceIdentifiers.Validate()
}
