package xattr

import "fmt"

const (
	capabilities_header = "SCHILY.xattr.security.capability"
)

var cap_empty_bitmask = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var supported_capabilities = map[string][]byte{
	"cap_net_bind_service": {1, 0, 0, 2, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

func AddCapabilities(pax map[string]string, capabilities []string) error {
	for _, cap := range capabilities {
		if _, supported := supported_capabilities[cap]; !supported {
			return fmt.Errorf("requested capability '%s' is not supported", cap)
		}
		if _, exists := pax[capabilities_header]; !exists {
			pax[capabilities_header] = string(cap_empty_bitmask)
		}
		val := []byte(pax[capabilities_header])
		for i, b := range supported_capabilities[cap] {
			val[i] = val[i] | b
		}
		pax[capabilities_header] = string(val)
	}
	return nil
}
