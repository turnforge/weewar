package fsbe

import (
	"path/filepath"

	"github.com/panyam/goutils/utils"
)

const WEEWAR_DATA_ROOT = "~/dev-app-data/weewar"

// For dev
func DevDataPath(path string) string {
	return filepath.Join(utils.ExpandUserPath(WEEWAR_DATA_ROOT), path)
}
