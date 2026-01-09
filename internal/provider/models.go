// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains model types for BCM Terraform provider resources.
// These models represent the Terraform state structure for various BCM entities.
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// EtcdCluster Models (bcm_cmetcd_cluster resource)
// =============================================================================

// EtcdClusterResourceModel represents the Terraform state for bcm_cmetcd_cluster resource.
// Maps to BCM CMEtcd service EtcdCluster entity.
//
// BCM API Field Mappings:
//
//	Terraform Schema       BCM API Field
//	-----------------      ---------------
//	id                  → uuid (identifier)
//	uuid                → uuid
//	name                → name
//	heartbeat_interval  → heartbeatInterval
//	election_timeout    → electionTimeout
//	options             → options (JSON-encoded string)
//	creation_time       → creationTime (computed)
//	revision_id         → revisionID (computed)
type EtcdClusterResourceModel struct {
	ID                types.String `tfsdk:"id"`
	UUID              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	HeartbeatInterval types.Int64  `tfsdk:"heartbeat_interval"`
	ElectionTimeout   types.Int64  `tfsdk:"election_timeout"`
	Options           types.String `tfsdk:"options"`
	CreationTime      types.Int64  `tfsdk:"creation_time"`
	RevisionID        types.Int64  `tfsdk:"revision_id"`
}

// =============================================================================
// KubeCluster Aligned Models (bcm_cmkube_cluster resource - aligned with BCM API)
// =============================================================================

// KubeClusterAlignedResourceModel represents the Terraform state for bcm_cmkube_cluster resource
// aligned with the actual BCM CMKube service KubeCluster entity.
//
// BCM API Field Mappings:
//
//	Terraform Schema                   BCM API Field
//	-----------------                  ---------------
//	id                              → uuid (identifier)
//	uuid                            → uuid
//	name                            → name
//	internal_network                → internalNetwork (UUID reference)
//	service_network                 → serviceNetwork (UUID reference)
//	pod_network                     → podNetwork (UUID reference)
//	etcd_cluster                    → etcdCluster (UUID reference)
//	version                         → version
//	pod_network_node_mask           → podNetworkNodeMask
//	kube_dns_ip                     → kubeDnsIp
//	kubernetes_api_server           → kubernetesApiServer
//	kubernetes_api_server_proxy_port → kubernetesApiServerProxyPort
//	trusted_domains                 → trustedDomains (list)
//	ingress_proxy_enable            → ingressProxyEnable
//	ingress_proxy_listen_port       → ingressProxyListenPort
//	ingress_proxy_backend_port      → ingressProxyBackendPort
//	options                         → options (JSON-encoded string)
//	app_groups                      → appGroups (nested block)
//	label_sets                      → labelSets (nested block)
//	users                           → users (nested block)
//	creation_time                   → creationTime (computed)
//	revision_id                     → revisionID (computed)
type KubeClusterAlignedResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	UUID                         types.String `tfsdk:"uuid"`
	Name                         types.String `tfsdk:"name"`
	InternalNetwork              types.String `tfsdk:"internal_network"`
	ServiceNetwork               types.String `tfsdk:"service_network"`
	PodNetwork                   types.String `tfsdk:"pod_network"`
	EtcdCluster                  types.String `tfsdk:"etcd_cluster"`
	Version                      types.String `tfsdk:"version"`
	PodNetworkNodeMask           types.String `tfsdk:"pod_network_node_mask"`
	KubeDnsIP                    types.String `tfsdk:"kube_dns_ip"`
	KubernetesAPIServer          types.String `tfsdk:"kubernetes_api_server"`
	KubernetesAPIServerProxyPort types.Int64  `tfsdk:"kubernetes_api_server_proxy_port"`
	TrustedDomains               types.List   `tfsdk:"trusted_domains"`
	IngressProxyEnable           types.Bool   `tfsdk:"ingress_proxy_enable"`
	IngressProxyListenPort       types.Int64  `tfsdk:"ingress_proxy_listen_port"`
	IngressProxyBackendPort      types.Int64  `tfsdk:"ingress_proxy_backend_port"`
	Options                      types.String `tfsdk:"options"`
	AppGroups                    types.List   `tfsdk:"app_groups"`
	LabelSets                    types.List   `tfsdk:"label_sets"`
	Users                        types.List   `tfsdk:"users"`
	CreationTime                 types.Int64  `tfsdk:"creation_time"`
	RevisionID                   types.Int64  `tfsdk:"revision_id"`
}

// KubeAppGroupModel represents a KubeAppGroup nested block in bcm_cmkube_cluster.
// Maps to BCM KubeAppGroup entity embedded in KubeCluster.appGroups[].
type KubeAppGroupModel struct {
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Applications types.List   `tfsdk:"applications"` // List of KubeAppModel
}

// KubeAppModel represents a KubeApp nested block within KubeAppGroup.
// Maps to BCM KubeApp entity embedded in KubeAppGroup.applications[].
type KubeAppModel struct {
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Manifest types.String `tfsdk:"manifest"`
}

// KubeLabelSetModel represents a KubeLabelSet nested block in bcm_cmkube_cluster.
// Maps to BCM KubeLabelSet entity embedded in KubeCluster.labelSets[].
type KubeLabelSetModel struct {
	Name   types.String `tfsdk:"name"`
	Labels types.Map    `tfsdk:"labels"` // Map of string to string
}

// KubeUserModel represents a KubeUser nested block in bcm_cmkube_cluster.
// Maps to BCM KubeUser entity embedded in KubeCluster.users[].
type KubeUserModel struct {
	Name   types.String `tfsdk:"name"`
	Groups types.List   `tfsdk:"groups"` // List of strings
}

// =============================================================================
// Device Role Models (nested blocks in bcm_cmdevice_device resource)
// =============================================================================

// KubeletRoleModel represents a kubelet_role nested block in bcm_cmdevice_device.
// Maps to BCM Role entity with childType="KubeletRole" in Device.roles[].
//
// BCM API Field Mappings:
//
//	Terraform Schema             BCM API Field
//	-----------------            ---------------
//	uuid                      → uuid (computed)
//	kube_cluster              → kubeCluster (UUID reference)
//	control_plane             → controlPlane
//	worker                    → worker
//	container_runtime_service → containerRuntimeService
//	max_pods                  → maxPods
//	options                   → options (JSON-encoded string)
//	custom_yaml               → customYaml
type KubeletRoleModel struct {
	UUID                    types.String `tfsdk:"uuid"`
	KubeCluster             types.String `tfsdk:"kube_cluster"`
	ControlPlane            types.Bool   `tfsdk:"control_plane"`
	Worker                  types.Bool   `tfsdk:"worker"`
	ContainerRuntimeService types.String `tfsdk:"container_runtime_service"`
	MaxPods                 types.Int64  `tfsdk:"max_pods"`
	Options                 types.String `tfsdk:"options"`
	CustomYAML              types.String `tfsdk:"custom_yaml"`
}

// EtcdHostRoleModel represents an etcd_host_role nested block in bcm_cmdevice_device.
// Maps to BCM Role entity with childType="EtcdHostRole" in Device.roles[].
//
// BCM API Field Mappings:
//
//	Terraform Schema           BCM API Field
//	-----------------          ---------------
//	uuid                    → uuid (computed)
//	etcd_cluster            → etcdCluster (UUID reference)
//	member_name             → memberName
//	spool                   → spool
//	listen_client_urls      → listenClientUrls (list)
//	listen_peer_urls        → listenPeerUrls (list)
//	advertise_client_urls   → advertiseClientUrls (list)
//	advertise_peer_urls     → advertisePeerUrls (list)
//	snapshot_count          → snapshotCount
//	max_snapshots           → maxSnapshots
type EtcdHostRoleModel struct {
	UUID                types.String `tfsdk:"uuid"`
	EtcdCluster         types.String `tfsdk:"etcd_cluster"`
	MemberName          types.String `tfsdk:"member_name"`
	Spool               types.String `tfsdk:"spool"`
	ListenClientURLs    types.List   `tfsdk:"listen_client_urls"`
	ListenPeerURLs      types.List   `tfsdk:"listen_peer_urls"`
	AdvertiseClientURLs types.List   `tfsdk:"advertise_client_urls"`
	AdvertisePeerURLs   types.List   `tfsdk:"advertise_peer_urls"`
	SnapshotCount       types.Int64  `tfsdk:"snapshot_count"`
	MaxSnapshots        types.Int64  `tfsdk:"max_snapshots"`
}
