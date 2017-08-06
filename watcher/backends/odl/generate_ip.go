package odl

import (
	"net"
	"sync"
)

// set initial to 10.0.0.2
var ip int32 = 10<<24 + 2
var mtx sync.Mutex

// Temporary method for generating IP addresses in a flat space
func generateIP() net.IP {

	// generate new address and increment for next call
	mtx.Lock()
	tmpip := ip
	ip++
	mtx.Unlock()

	// convert to net.IP
	octet4 := byte(tmpip)
	octet3 := byte(tmpip >> 8)
	octet2 := byte(tmpip >> 16)
	octet1 := byte(tmpip >> 24)
	return net.IPv4(octet1, octet2, octet3, octet4)
}
