# API Research: bcm_cmkube_cluster etcd_nodes

**Date**: 2025-11-25
**Issue**: [#25](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/25)
**Script**: `sampleRest/cmkube-etcd-test.py`

## Research Questions & Findings

### RQ-001: Does BCM cmkube API support `etcdNodes` field?

**Answer**: YES

BCM API successfully accepts `etcdNodes` field in addKubeCluster request.

```json
// Request
{
  "service": "cmkube",
  "call": "addKubeCluster",
  "args": [
    {
      "baseType": "KubeCluster",
      "uuid": "client-generated-uuid",
      "name": "cluster-name",
      "masterNodes": ["master-node-uuid"],
      "etcdNodes": ["etcd-node-1", "etcd-node-2", "etcd-node-3"]
    },
    false
  ]
}

// Response
{
  "success": true,
  "task_uuid": "00000000-0000-0000-0000-000000000000",
  "updated_entity": null,
  "validation": []
}
```

### RQ-002: Field name mapping

**Answer**: `etcdNodes` (camelCase in BCM API)

Terraform attribute: `etcd_nodes` (snake_case)

### RQ-003: Does getKubeCluster return etcdNodes?

**Answer**: NO

The `etcdNodes` field is NOT returned in getKubeCluster response. This is consistent with `masterNodes` and `workerNodes` behavior.

**Implementation Impact**: Must preserve `etcd_nodes` from plan/state during Read operations.

### RQ-004: Does BCM enforce odd number requirement for etcd nodes?

**Answer**: NOT TESTED

Test environment only has 2 nodes. Further testing needed with 2 or 4 etcd nodes to verify if BCM enforces odd numbers.

### RQ-005: Empty etcdNodes vs omitted?

**Answer**: NOT TESTED

When etcdNodes is omitted, BCM defaults to running etcd on master nodes.

## Attribute Support Matrix

| Terraform Attribute | BCM Field | API Accept | API Return | Support Level |
|---------------------|-----------|------------|------------|---------------|
| etcd_nodes | etcdNodes | YES | NO | WRITE_ONLY |
| version | version | YES | YES | FULL_SUPPORT |
| cni_plugin | cniPlugin | YES | NO | WRITE_ONLY |
| dns_servers | dnsServers | YES | NO | WRITE_ONLY |
| overlay_network | overlayNetwork | YES | NO | WRITE_ONLY |
| load_balancer_mode | loadBalancerMode | YES | NO | WRITE_ONLY |
| storage_classes | storageClasses | YES | NO | WRITE_ONLY |
| addons | addons | YES | NO | WRITE_ONLY |
| ingress_controller | ingressController | YES | NO | WRITE_ONLY |
| master_nodes | masterNodes | YES | NO | WRITE_ONLY |
| worker_nodes | workerNodes | YES | NO | WRITE_ONLY |

## getKubeCluster Response Structure

Fields returned by BCM API:

```json
{
  "appGroups": [],
  "baseType": "KubeCluster",
  "capiNamespace": "default",
  "capiTemplate": false,
  "childType": "",
  "etcdCluster": "00000000-0000-0000-0000-000000000000",
  "external": false,
  "externalIngressServer": "",
  "externalPort": 0,
  "extra_values": null,
  "ingressProxyBackendPort": 0,
  "ingressProxyEnable": false,
  "ingressProxyListenPort": 443,
  "internalNetwork": "00000000-0000-0000-0000-000000000000",
  "kubeCluster": "00000000-0000-0000-0000-000000000000",
  "kubeDnsIp": "0.0.0.0",
  "kubeadm_ca_cert": "",
  "kubeadm_ca_key": "",
  "kubeadm_init_cert_key": "",
  "kubeadm_init_file": "",
  "kubernetesApiServer": "",
  "kubernetesApiServerProxyPort": 6444,
  "labelSets": [],
  "modified": false,
  "moduleFileTemplate": "",
  "name": "cluster-name",
  "notes": "",
  "options": null,
  "podNetwork": "00000000-0000-0000-0000-000000000000",
  "podNetworkNodeMask": "",
  "revision": "",
  "serviceNetwork": "00000000-0000-0000-0000-000000000000",
  "to_be_removed": false,
  "trustedDomains": [],
  "users": [],
  "uuid": "cluster-uuid",
  "version": "1.28.0"
}
```

**Notable Finding**: `etcdCluster` UUID is returned, but `etcdNodes` list is NOT returned.

## Test Environment

- BCM Endpoint: https://172.21.15.254:8081
- Available Nodes: 2 (limits etcd testing to 1-2 nodes)
- Node UUIDs:
  - Node 0: 9f885869-a146-4cd6-af1f-f9b6c674a84c
  - Node 1: 95a75a76-09ba-4bbe-9ff0-ae68dc0c330e

## Implementation Recommendations

1. **etcd_nodes attribute**: Add as optional list of strings
2. **Schema description**: Document NVIDIA recommendation (3 nodes for HA)
3. **Write-only handling**: Preserve from plan/state during Read
4. **ImportStateVerifyIgnore**: Add etcd_nodes to ignore list
5. **Test flexibility**: Support tests with 1-3 etcd nodes based on available nodes

## References

- Research Script: `sampleRest/cmkube-etcd-test.py`
- Test Output: `sampleRest/cmkube-etcd-test-output.json`
- Existing Implementation: `internal/provider/resource_cmkube_cluster.go`
