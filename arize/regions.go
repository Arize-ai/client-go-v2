package arize

import "github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"

// Region represents an Arize deployment region.
type Region = sdkconfig.Region

// RegionEndpoints holds the per-region endpoint defaults.
type RegionEndpoints = sdkconfig.RegionEndpoints

const (
	RegionUSCentral = sdkconfig.RegionUSCentral
	RegionEUWest    = sdkconfig.RegionEUWest
	RegionCACentral = sdkconfig.RegionCACentral
	RegionUSEast    = sdkconfig.RegionUSEast
)

// RegionEndpointsFor returns the default endpoint configuration for r.
// The second return is false if r is not a known region.
func RegionEndpointsFor(r Region) (RegionEndpoints, bool) {
	return sdkconfig.RegionEndpointsFor(r)
}
