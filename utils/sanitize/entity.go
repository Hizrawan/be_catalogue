package sanitize

import "strings"

func ID(id string) string {
	return strings.Replace(id, ":", "_", -1)
}
