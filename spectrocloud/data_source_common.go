package spectrocloud

import (
	"bytes"
	"fmt"
	"sort"
)

func toDatasourcesId(prefix string, labels map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString(prefix)

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("-%s-%s", k, labels[k]))
	}

	id := buf.String()
	return id
}
