# Quick Start: BCM Network Resource

## Build & Test
```bash
make install
export TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!"
go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork
```

## Example Usage
```hcl
resource "bcm_cmnet_network" "example" {
  name              = "compute-network"
  subnet            = "10.0.1.0/24"
  gateway           = "10.0.1.1"
  mtu               = 9000
  domain_name       = "cluster.local"
  dhcp_range_start  = "10.0.1.100"
  dhcp_range_end    = "10.0.1.200"
}
```

## Implementation Files
- `/workspace/internal/provider/resource_cmnet_network.go` - Resource implementation
- `/workspace/internal/provider/resource_cmnet_network_test.go` - Tests
- `/workspace/examples/resources/bcm_cmnet_network/` - Examples
