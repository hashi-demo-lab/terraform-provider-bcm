## 0.1.0 (Unreleased)

FEATURES:

* **New Resource:** `bcm_cmdevice_category` - Manage BCM device categories
* **New Resource:** `bcm_cmdevice_device` - Manage BCM devices
* **New Resource:** `bcm_cmpart_softwareimage` - Manage BCM software images
* **New Resource:** `bcm_cmnet_network` - Manage BCM networks
* **New Resource:** `bcm_cmkube_cluster` - Manage BCM Kubernetes clusters
* **New Resource:** `bcm_cmuser_user` - Manage BCM users
* **New Data Source:** `bcm_cmdevice_categories` - Query BCM categories
* **New Data Source:** `bcm_cmdevice_nodes` - Query BCM nodes
* **New Data Source:** `bcm_cmdevice_roles` - Query BCM roles
* **New Data Source:** `bcm_cmnet_networks` - Query BCM networks
* **New Data Source:** `bcm_cmpart_softwareimages` - Query BCM software images
* **New Data Source:** `bcm_cmpart_partitions` - Query BCM partitions
* **New Data Source:** `bcm_cmkube_clusters` - Query BCM Kubernetes clusters
* **New Data Source:** `bcm_cmuser_users` - Query BCM users
* **New Action:** `bcm_cmdevice_power` - Control device power state

BUG FIXES:

* resource/bcm_cmdevice_category: Fixed roles[].uuid computed value never being populated from BCM API response ([#83](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/83))
