# BCM Category API Schema Documentation

Generated: 2025-11-21 07:06:29

Total categories analyzed: 1

## Category Attributes

| Attribute | Type | Nullable | Examples |
|-----------|------|----------|----------|
| `accessSettings` | NoneType | Yes |  |
| `allowNetworkingRestart` | bool | No | False |
| `authenticationService` | str | No | AUTO |
| `baseType` | str | No | Category |
| `biosSetup` | NoneType | Yes |  |
| `bmcSettings` | dict | No | {'baseType': 'BMCSettings', 'childType': '', 'e... |
| `bootLoader` | str | No | SYSLINUX |
| `bootLoaderFile` | str | No |  |
| `bootLoaderProtocol` | str | No | HTTP |
| `childType` | str | No |  |
| `dataNode` | bool | No | False |
| `defaultGateway` | str | No | 0.0.0.0 |
| `defaultGatewayMetric` | int | No | 0 |
| `disksetup` | str | No | <?xml version="1.0" encoding="UTF-8"?>

<diskSe... |
| `dpuSettings` | NoneType | Yes |  |
| `excludeListFull` | str | No | # For details on the exclude patterns defined h... |
| `excludeListGrab` | str | No | - /.autofsck
- /boot/efi
- /boot/grub/device.ma... |
| `excludeListGrabnew` | str | No | - /.autofsck
- /boot/efi
- /cgroup/*
- /cm/imag... |
| `excludeListManipulateScript` | str | No |  |
| `excludeListSync` | str | No | # For details on the exclude patterns defined h... |
| `excludeListUpdate` | str | No | # For details on the exclude patterns defined h... |
| `extra_values` | NoneType | Yes |  |
| `finalize` | str | No |  |
| `fips` | str | No | NO |
| `fsexports` | list | No | [] |
| `fsmounts` | list | No | [{'baseType': 'FSMount', 'childType': '', 'devi... |
| `gpuSettings` | list | No | [] |
| `initialize` | str | No |  |
| `installBootRecord` | bool | No | False |
| `installMode` | str | No | AUTO |
| `interactiveUser` | str | No | ALWAYS |
| `ioScheduler` | str | No |  |
| `kernelOutputConsole` | str | No |  |
| `kernelParameters` | str | No |  |
| `kernelVersion` | str | No |  |
| `managementNetwork` | str | No | 84d8d82b-3ae7-4433-a793-bb44d5c3b4fe |
| `modified` | bool | No | False |
| `modules` | list | No | [] |
| `name` | str | No | default |
| `nameServers` | list | No | [] |
| `newNodeInstallMode` | str | No | FULL |
| `nodeInstallerDisk` | bool | No | False |
| `notes` | str | No |  |
| `parent_uuid` | str | No | 0ae6d733-3015-4479-bfab-ce2d237a2809 |
| `proxySettings` | NoneType | Yes |  |
| `raidconf` | str | No |  |
| `revision` | str | No |  |
| `roles` | list | No | [] |
| `seLinuxSettings` | NoneType | Yes |  |
| `searchDomains` | list | No | [] |
| `services` | list | No | [] |
| `softwareImageProxy` | dict | No | {'baseType': 'SoftwareImageProxy', 'childType':... |
| `staticRoutes` | list | No | [] |
| `timeServers` | list | No | [] |
| `timeZoneSettings` | NoneType | Yes |  |
| `to_be_removed` | bool | No | False |
| `useExclusivelyFor` | str | No |  |
| `uuid` | str | No | 0ae6d733-3015-4479-bfab-ce2d237a2809 |
| `versionConfigFiles` | bool | No | False |
| `ztpSettings` | NoneType | Yes |  |

## Full Category Example

```json
{
  "accessSettings": null,
  "allowNetworkingRestart": false,
  "authenticationService": "AUTO",
  "baseType": "Category",
  "biosSetup": null,
  "bmcSettings": {
    "baseType": "BMCSettings",
    "childType": "",
    "extraArguments": "",
    "extra_values": null,
    "firmwareManageMode": "AUTO",
    "leakPolicy": "NONE",
    "leakReactionDelay": 0.0,
    "modified": false,
    "password": "6wMiLy4I",
    "powerResetDelay": 0,
    "privilege": "ADMINISTRATOR",
    "revision": "",
    "to_be_removed": false,
    "userID": 4,
    "userName": "bright",
    "uuid": "3c0cf401-f2dd-4d56-9387-f84dac6bd9fa"
  },
  "bootLoader": "SYSLINUX",
  "bootLoaderFile": "",
  "bootLoaderProtocol": "HTTP",
  "childType": "",
  "dataNode": false,
  "defaultGateway": "0.0.0.0",
  "defaultGatewayMetric": 0,
  "disksetup": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n\n<diskSetup>\n  <device>\n    <blockdev>/dev/sda</blockdev>\n    <blockdev>/dev/hda</blockdev>\n    <blockdev>/dev/vda</blockdev>\n    <blockdev>/dev/xvda</blockdev>\n    <blockdev>/dev/nvme0n1</blockdev>\n    <blockdev mode=\"cloud\">/dev/sdb</blockdev>\n    <blockdev mode=\"cloud\">/dev/hdb</blockdev>\n    <blockdev mode=\"cloud\">/dev/vdb</blockdev>\n    <blockdev mode=\"cloud\">/dev/xvdf</blockdev>\n    <partition id=\"a0\" partitiontype=\"esp\">\n      <size>100M</size>\n      <type>linux</type>\n      <filesystem>fat</filesystem>\n      <mountPoint>/boot/efi</mountPoint>\n      <mountOptions>defaults,noatime,nodiratime</mountOptions>\n    </partition>\n    <partition id=\"a1\">\n      <size>20G</size>\n      <type>linux</type>\n      <filesystem>xfs</filesystem>\n      <mountPoint>/</mountPoint>\n      <mountOptions>defaults,noatime,nodiratime</mountOptions>\n    </partition>\n    <partition id=\"a2\">\n      <size>6G</size>\n      <type>linux</type>\n      <filesystem>xfs</filesystem>\n      <mountPoint>/var</mountPoint>\n      <mountOptions>defaults,noatime,nodiratime</mountOptions>\n    </partition>\n    <partition id=\"a3\">\n      <size>2G</size>\n      <type>linux</type>\n      <filesystem>xfs</filesystem>\n      <mountPoint>/tmp</mountPoint>\n      <mountOptions>defaults,noatime,nodiratime,nosuid,nodev</mountOptions>\n    </partition>\n    <partition id=\"a4\">\n      <size>12G</size>\n      <type>linux swap</type>\n    </partition>\n    <partition id=\"a5\">\n      <size>max</size>\n      <type>linux</type>\n      <filesystem>xfs</filesystem>\n      <mountPoint>/local</mountPoint>\n      <mountOptions>defaults,noatime,nodiratime</mountOptions>\n    </partition>\n  </device>\n</diskSetup>\n",
  "dpuSettings": null,
  "excludeListFull": "# For details on the exclude patterns defined here please refer to\n# the FILTER RULES section of the rsync man page.\n#\n# Files that match these patterns will not be installed onto the node.\n- lost+found/\n- /proc/*\n- /sys/*\n- /boot/efi",
  "excludeListGrab": "- /.autofsck\n- /boot/efi\n- /boot/grub/device.map\n- /boot/grub/grub.conf\n- /boot/grub/menu.lst\n- /boot/grub*/device.map\n- /boot/grub*/grub.cfg\n- /boot/grub*/grubenv\n- /cgroup/*\n- /cm/images/*\n- /cm/local/apps/cmd/etc/cert.*\n- /cm/local/apps/openldap/etc/certs/ldap.key\n- /cm/local/apps/openldap/etc/certs/ldap.pem\n- /cm/node-installer/*\n- /cm/node-installer-ebs/*\n- /cm/shared/*\n- /data/*\n- /dev/*\n- /etc/aliases.db\n- /etc/blkid/*\n- /etc/exports\n- /etc/fstab\n- /etc/HOSTNAME\n- /etc/hostname\n- /etc/hosts\n- /etc/lvm/cache/*\n- /etc/lvm/archive/*\n- /etc/lvm/backup/*\n- /etc/mtab\n- /etc/ntp.conf\n- /etc/ntp/*\n- /etc/openvpn/*\n- /etc/postfix/main.cf\n- /etc/resolv.conf\n- /etc/sysconfig/network\n- /etc/sysconfig/network-scripts/ifcfg-eth*\n- /etc/sysconfig/network-scripts/ifcfg-em*\n- /etc/sysconfig/network-scripts/ifcfg-eno*\n- /etc/sysconfig/network-scripts/ifcfg-enp*\n- /etc/sysconfig/network-scripts/ifcfg-ens*\n- /etc/sysconfig/network-scripts/ifcfg-enx*\n- /etc/sysconfig/network-scripts/ifcfg-ib*\n- /etc/sysconfig/network-scripts/ifcfg-br*\n- /etc/sysconfig/network-scripts/ifcfg-bond*\n- /etc/sysconfig/network-scripts/ifcfg-p*\n- /etc/network/interfaces.d/*\n- /etc/udev/rules.d/*-persistent-*\n- /fhgfs/*\n- /local/*\n- lost+found/\n- /media/*\n- /proc/*\n- /run/*\n- /scratch/*\n- /sys/*\n- /tftpboot/*\n- /tmp/*\n- /var/cache/yum/*\n- /var/cache/dnf/*\n- /var/lib/dhclient/*\n- /var/lib/dhcpcd/*\n- /var/lib/ldap/*\n- /var/lib/logrotate.status\n- /var/lib/mlocate/*\n- /var/lib/nfs/*\n- /var/lib/ntp/drift\n- /var/lib/ntp/proc/*\n- /var/lib/rpm/__db.*\n- /var/lib/sss/*\n- /var/lib/systemd/random-seed\n- /var/lock/subsys/*\n- /var/log/*\n- /var/run/*.pid\n- /var/run/*/*.pid\n- /var/spool/anacron/*\n- /var/spool/cmd/*\n- /var/spool/mail/*\n- /var/spool/postfix/*\n- /var/tmp/*",
  "excludeListGrabnew": "- /.autofsck\n- /boot/efi\n- /cgroup/*\n- /cm/images/*\n- /cm/local/apps/cmd/etc/cert.*\n- /cm/local/apps/openldap/etc/certs/ldap.key\n- /cm/local/apps/openldap/etc/certs/ldap.pem\n- /cm/node-installer/*\n- /cm/node-installer-ebs/*\n- /cm/shared/*\n- /dev/.udev/*\n- /dev/bus/*\n- /dev/cpu/*\n- /dev/disk/by-id/*\n- /dev/disk/by-path/*\n- /dev/disk/by-uuid/*\n- /dev/infiniband/*\n- /dev/shm/*\n- /dev/usbdev*\n- /etc/HOSTNAME\n- /etc/hostname\n- /etc/lvm/cache/*\n- /etc/lvm/archive/*\n- /etc/lvm/backup/*\n- /etc/openvpn/*\n- /etc/resolv.conf\n- /etc/sysconfig/network-scripts/ifcfg-eth*\n- /etc/sysconfig/network-scripts/ifcfg-em*\n- /etc/sysconfig/network-scripts/ifcfg-eno*\n- /etc/sysconfig/network-scripts/ifcfg-enp*\n- /etc/sysconfig/network-scripts/ifcfg-ens*\n- /etc/sysconfig/network-scripts/ifcfg-enx*\n- /etc/sysconfig/network-scripts/ifcfg-ib*\n- /etc/sysconfig/network-scripts/ifcfg-br*\n- /etc/sysconfig/network-scripts/ifcfg-bond*\n- /etc/sysconfig/network-scripts/ifcfg-p*\n- /etc/network/interfaces.d/*\n- /etc/udev/rules.d/*-persistent-*\n- lost+found/\n- /media/*\n- /proc/*\n- /run/*\n- /sys/*\n- /tftpboot/*\n- /tmp/*\n- /var/cache/yum/*\n- /var/cache/dnf/*\n- /var/lib/dhclient/*\n- /var/lib/dhcpcd/*\n- /var/lib/logrotate.status\n- /var/lib/mlocate/*\n- /var/lib/ntp/proc/*\n- /var/lib/systemd/random-seed\n- /var/spool/anacron/*\n- /var/spool/cmd/*\n- /var/log/node-installer\n- /var/tmp/*\n- /var/lib/sss/*",
  "excludeListManipulateScript": "",
  "excludeListSync": "# For details on the exclude patterns defined here please refer to\n# the FILTER RULES section of the rsync man page.\n#\n# Files that exist on a node and match one of these patterns will not be\n# modified or deleted. Any files that match one of these patterns and that\n# exist in the image but are absent on the node, will be copied to the node.\n- /.autofsck\n- /boot/grub*/grub.cfg\n- /cm/local/apps/openldap/etc/certs/ldap.key\n- /cm/local/apps/openldap/etc/certs/ldap.pem\n- /data/*\n- /home/*\n- /local/*\n- /scratch/*\n- /tmp/*\n- /var/log/*\n- /var/tmp/*\n- /var/spool/mail/*\n- /var/spool/postfix/*\n\n# NVidia drivers (cuda-driver)\n- /lib/modules/*/kernel/drivers/video/nvidia*.ko\n- /usr/lib/modules/*/kernel/drivers/video/nvidia*.ko\n\n# Files that exist on a node and match one of these patterns will not be\n# modified or deleted. Any files that match one of these patterns and that\n# exist in the image will never be copied to the node.\n# (The prefix \"no-new-files: \" will be removed before passing to rsync.)\nno-new-files: - /boot/efi\nno-new-files: - /cm/images\nno-new-files: - /cm/shared/*\nno-new-files: - lost+found/\nno-new-files: - /proc/*\nno-new-files: - /sys/*\nno-new-files: - /tftpboot/*\nno-new-files: - /.autorelabel\nno-new-files: - /var/lib/logrotate.status\nno-new-files: - /var/lib/sss/*\nno-new-files: - /var/lib/systemd/random-seed\nno-new-files: - /var/spool/anacron/*",
  "excludeListUpdate": "# For details on the exclude patterns defined here please refer to\n# the FILTER RULES section of the rsync man page.\n#\n# Files that exist on a node and match one of these patterns will not be\n# modified or deleted. Any files that match one of these patterns and that\n# exist in the image but are absent on the node, will be copied to the node.\n- /.autofsck\n- /.autorelabel\n- /boot/boot\n- /boot/grub/device.map\n- /boot/grub/grub.conf\n- /boot/grub/menu.lst\n- /boot/grub*/device.map\n- /boot/grub*/fonts\n- /boot/grub*/grub.cfg\n- /boot/grub*/grubenv\n- /boot/grub*/i386-pc\n- /boot/grub*/locale\n- /boot/grub2\n- /boot/initrd-*.orig\n- /cm/local/apps/cmd/etc/*\n- /cm/local/apps/openldap/etc/certs/ldap.key\n- /cm/local/apps/openldap/etc/certs/ldap.pem\n- /data/*\n- /etc/aliases.db\n- /etc/blkid/*\n- /etc/cm/burnrc\n- /etc/dhcpd.*\n- /etc/exports\n- /etc/fstab\n- /etc/hosts\n- /etc/HOSTNAME\n- /etc/hostname\n- /etc/lvm/cache/.cache\n- /etc/lvm/archive/*\n- /etc/lvm/backup/*\n- /etc/mtab\n- /etc/ntp.conf\n- /etc/ntp/step-tickers\n- /etc/openvpn\n- /etc/pam.d/sshd\n- /etc/postfix/main.cf\n- /etc/rc.d/rc*.d/*dhcpd\n- /etc/rc.d/rc*.d/*maui\n- /etc/rc.d/rc*.d/*moab\n- /etc/rc.d/rc*.d/*munge\n- /etc/rc.d/rc*.d/*nfs\n- /etc/rc.d/rc*.d/*opensmd\n- /etc/rc.d/rc*.d/*opensm\n- /etc/rc.d/rc*.d/*pbs_mom\n- /etc/rc.d/rc*.d/*pbs\n- /etc/rc.d/rc*.d/*portmap\n- /etc/rc.d/rc*.d/*rpcbind\n- /etc/rc.d/rc*.d/*sgemaster.sge1\n- /etc/rc.d/rc*.d/*sgeexecd\n- /etc/rc.d/rc*.d/*slurm\n- /etc/rc.d/rc*.d/*slurmdbd\n- /etc/reader.conf\n- /etc/resolv.conf\n- /etc/security/pam_bright.d/cm-check-alloc.conf\n- /etc/sensors3.conf\n- /etc/sysconfig/network\n- /etc/sysconfig/network-scripts/ifcfg-*\n- /etc/network/interfaces.d/*\n- /etc/sysconfig/openib\n- /etc/systemd/system/*.wants/dhcpd.service\n- /etc/systemd/system/*.wants/maui.service\n- /etc/systemd/system/*.wants/moab.service\n- /etc/systemd/system/*.wants/munge.service\n- /etc/systemd/system/*.wants/nfs.service\n- /etc/systemd/system/*.wants/opensmd.service\n- /etc/systemd/system/*.wants/opensm.service\n- /etc/systemd/system/*.wants/pbs_mom.service\n- /etc/systemd/system/*.wants/pbs.service\n- /etc/systemd/system/*.wants/portmap.service\n- /etc/systemd/system/*.wants/rpcbind.service\n- /etc/systemd/system/*.wants/sgemaster.sge1.service\n- /etc/systemd/system/*.wants/sgeexecd.service\n- /etc/systemd/system/*.wants/slurmctld.service\n- /etc/systemd/system/*.wants/slurmd.service\n- /etc/systemd/system/*.wants/slurmdbd.service\n- /fhgfs/*\n- /home/*\n- /local/*\n- /mnt/*\n- /root/.bash_history\n- /root/.modulesbeginenv\n- /root/.ssh/known_hosts\n- /scratch/*\n- /tmp/*\n- /var/cache/man/*\n- /var/empty/*\n- /var/lib/dhclient/*\n- /var/lib/dhcp/*\n- /var/lib/dhcpcd/*\n- /var/lib/gssproxy/default.sock\n- /var/lib/logrotate.status\n- /var/lib/misc/postfix.aliasesdb-stamp\n- /var/lib/mlocate/*\n- /var/lib/nfs/*\n- /var/lib/ntp/drift\n- /var/lib/ntp/proc\n- /var/lib/plymouth/boot-duration\n- /var/lib/postfix/master.lock\n- /var/lib/random-seed\n- /var/log/*\n- /var/net-snmp*\n- /var/spool/*\n- /var/tmp/*\n\n# OFED\n- /usr/sbin/ibpd\n- /etc/infiniband\n- /usr/bin/ibdev2netdev\n- /etc/modprobe.d/mlx4_en.conf\n- /etc/modprobe.d/ib*.conf\n- /etc/rc.d/*/*openibd\n- /etc/udev/rules.d/*-ibpd.rules\n- /etc/udev/rules.d/*-ib.rules\n- /etc/udev/rules.d/*-persistent-net.rules\n- /etc/udev/rules.d/*-persistent-cd.rules\n- /sbin/connectx_port_config\n- /sbin/sysctl_perf_tuning\n- /var/cache/sysctl_perf_tuning\n\n# NVidia drivers (cuda-driver)\n- /lib/modules/*/kernel/drivers/video/nvidia*.ko\n- /usr/lib/modules/*/kernel/drivers/video/nvidia*.ko\n\n# Files that exist on a node and match one of these patterns will not be\n# modified or deleted. Any files that match one of these patterns and that\n# exist in the image will never be copied to the node.\n# (The prefix \"no-new-files: \" will be removed before passing to rsync.)\nno-new-files: - /boot/efi\nno-new-files: - /cgroup/*\nno-new-files: - /cm/images\nno-new-files: - /cm/node-installer-ebs\nno-new-files: - /cm/shared/*\nno-new-files: - /dev/*\nno-new-files: - lost+found/\nno-new-files: - /media/*\nno-new-files: - /proc/*\nno-new-files: - /run/*\nno-new-files: - /sys/*\nno-new-files: - /tftpboot/*\nno-new-files: - /var/lock/*\nno-new-files: - /var/lib/ldap/*\nno-new-files: - /var/lib/rpm/__db.*\nno-new-files: - /var/lib/sss/*\nno-new-files: - /var/lib/systemd/random-seed\nno-new-files: - /var/run/*\nno-new-files: - /var/spool/anacron/*\nno-new-files: - /.autorelabel",
  "extra_values": null,
  "finalize": "",
  "fips": "NO",
  "fsexports": [],
  "fsmounts": [
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "none",
      "dump": false,
      "extra_values": null,
      "filesystem": "devpts",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "gid=5,mode=620",
      "mountpoint": "/dev/pts",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "6316e6d7-88b5-4cdd-892d-6e604df33f5e"
    },
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "none",
      "dump": false,
      "extra_values": null,
      "filesystem": "proc",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "defaults,nosuid",
      "mountpoint": "/proc",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "dcc2f713-1ab9-4deb-9be2-6093c3a25bad"
    },
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "none",
      "dump": false,
      "extra_values": null,
      "filesystem": "sysfs",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "defaults",
      "mountpoint": "/sys",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "d3f8bbc2-a459-4211-acbc-09679bd03b9d"
    },
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "none",
      "dump": false,
      "extra_values": null,
      "filesystem": "tmpfs",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "defaults",
      "mountpoint": "/dev/shm",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "ab293a44-ca9f-4a4c-9a08-de5a97d6647c"
    },
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "$localnfsserver:/cm/shared",
      "dump": false,
      "extra_values": null,
      "filesystem": "nfs",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "rsize=32768,wsize=32768,hard,async",
      "mountpoint": "/cm/shared",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "8d2c0ee5-52b9-4451-8ec8-5694d797bc31"
    },
    {
      "baseType": "FSMount",
      "childType": "",
      "device": "$localnfsserver:/home",
      "dump": false,
      "extra_values": null,
      "filesystem": "nfs",
      "fsck": "NONE",
      "modified": false,
      "mountoptions": "rsize=32768,wsize=32768,hard,async",
      "mountpoint": "/home",
      "rdma": false,
      "revision": "",
      "to_be_removed": false,
      "uuid": "3421849c-b740-4a6d-979b-29b9194f3860"
    }
  ],
  "gpuSettings": [],
  "initialize": "",
  "installBootRecord": false,
  "installMode": "AUTO",
  "interactiveUser": "ALWAYS",
  "ioScheduler": "",
  "kernelOutputConsole": "",
  "kernelParameters": "",
  "kernelVersion": "",
  "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
  "modified": false,
  "modules": [],
  "name": "default",
  "nameServers": [],
  "newNodeInstallMode": "FULL",
  "nodeInstallerDisk": false,
  "notes": "",
  "parent_uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
  "proxySettings": null,
  "raidconf": "",
  "revision": "",
  "roles": [],
  "seLinuxSettings": null,
  "searchDomains": [],
  "services": [],
  "softwareImageProxy": {
    "baseType": "SoftwareImageProxy",
    "childType": "",
    "extra_values": null,
    "modified": false,
    "parentSoftwareImage": "8482c4e9-383c-43de-873f-8c54ee77ee74",
    "revision": "",
    "revisionID": -1,
    "to_be_removed": false,
    "uuid": "7abe08d4-4c18-4d66-9eff-fa1a1b87e84c"
  },
  "staticRoutes": [],
  "timeServers": [],
  "timeZoneSettings": null,
  "to_be_removed": false,
  "useExclusivelyFor": "",
  "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
  "versionConfigFiles": false,
  "ztpSettings": null
}
```
