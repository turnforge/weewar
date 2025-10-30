package assets

import (
	_ "embed"
)

//go:embed weewar-rules.json
var RulesDataJSON []byte

//go:embed weewar-damage.json
var RulesDamageDataJSON []byte
