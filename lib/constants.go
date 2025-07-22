package weewar

import (
	"path/filepath"

	"github.com/panyam/goutils/utils"
)

const (
	DefaultTileWidth  = 64.0
	DefaultTileHeight = 64.0
	DefaultYIncrement = 48.0
	MaxUnits          = 20
)

const SQRT3 = 1.732050808 // sqrt(3)

const WEEWAR_DATA_ROOT = "~/dev-app-data/weewar"

// For dev
func DevDataPath(path string) string {
	return filepath.Join(utils.ExpandUserPath(WEEWAR_DATA_ROOT), path)
}
