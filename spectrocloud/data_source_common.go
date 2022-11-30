package spectrocloud

import (
	"bytes"
	"fmt"
)

func toDatasourcesId(prefix string, labels map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString(prefix)

	for k, v := range labels {
		buf.WriteString(fmt.Sprintf("-%s-%s", k, v))
	}

	id := buf.String()
	return id
}
