package templates

import (
	"strings"

	"github.com/caedis/noreza/internal/mapping"
)

func concatKeys(keys []mapping.KeyMapping) string {
	var keyVals []string

	for _, v := range keys {
		if v.Mode == mapping.Mouse {
			keyVals = append(keyVals, mapping.CodeToMouse[v.Code])
		} else {
			keyVals = append(keyVals, mapping.CodeToKeyFriendly[v.Code])
		}
	}

	return strings.Join(keyVals, "\n")
}
