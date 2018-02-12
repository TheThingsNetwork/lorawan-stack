// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import "encoding/gob"

func init() {
	gob.Register(&RxMetadata_EncryptedFineTimestamp{})
	gob.Register(&RxMetadata_FineTimestampValue{})
}
