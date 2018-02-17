// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scripting

import (
	"context"
)

// Engine represents a scripting engine.
type Engine interface {
	Run(ctx context.Context, script string, env map[string]interface{}) (interface{}, error)
}
