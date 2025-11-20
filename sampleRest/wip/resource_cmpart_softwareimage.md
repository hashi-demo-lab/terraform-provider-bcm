# example resource_cmpart_softwareimage the ars is the software image entity

POST https://172.21.15.254:8081/json
{
"service": "cmpart",
"call": "addSoftwareImage",
"args": [softwareImage, force]
}

{"service":"CMPart","call":"addSoftwareImage","args":[{"uuid":"eaad50d3-432a-4703-a9f8-66551c255a69","baseType":"SoftwareImage","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"cloned","path":"/cm/images/cloned","originalImage":"8482c4e9-383c-43de-873f-8c54ee77ee74","fileOperationInProgress":false,"kernelVersion":"6.8.0-51-generic","kernelParameters":"rd.driver.blacklist=nouveau","kernelOutputConsole":"tty0","creationTime":1763626149,"modules":[{"uuid":"16867c38-7b60-4fbe-b4fb-888fd3ea4ba5","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"aacraid","parameters":""},{"uuid":"a31f8c39-5964-4203-96f0-9ed1a9a33b09","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ahci","parameters":""},{"uuid":"e52b6cff-c573-47d3-9200-ab6981ad0b6a","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"aic79xx","parameters":""},{"uuid":"dc5ac94e-a6ee-46bb-a87e-852cb9ed18b7","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"aic7xxx","parameters":""},{"uuid":"fa09b5d7-7235-40e2-8fc7-f31b2b09de14","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"arcmsr","parameters":""},{"uuid":"62f426d6-511e-4cf7-a618-dd1434c9a827","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hpsa","parameters":""},{"uuid":"831d970d-eb2d-4c96-9f5c-662302caaf91","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"megaraid","parameters":""},{"uuid":"52e7933e-6633-4f33-a60a-375d73469d6e","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"megaraid_sas","parameters":""},{"uuid":"eb5cad38-a777-4aad-878f-c4fa19925fd7","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"mpt3sas","parameters":""},{"uuid":"e0dcf865-d754-4f73-af5d-d6f3b114345b","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"mptsas","parameters":""},{"uuid":"cb26a5c6-57c3-4e99-8667-400f81a6608c","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"mptscsih","parameters":""},{"uuid":"8f39c776-d814-429d-9d31-d42e75d074f4","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"mptspi","parameters":""},{"uuid":"58fc64ba-937c-49e3-a5cd-af92cf28d9bf","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"sata_nv","parameters":""},{"uuid":"9f9aa38b-d205-431f-8587-984dd1726ee0","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"sata_sil","parameters":""},{"uuid":"b73f9069-9e33-4ffe-b812-8151300b86e3","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"sata_svw","parameters":""},{"uuid":"e17cfe49-91e5-4ca4-a7ef-49efc30fd12d","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"bnx2","parameters":""},{"uuid":"2f286317-5603-4a2c-a5b1-65a8575d7087","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"bnx2x","parameters":""},{"uuid":"4b35cfc1-f43c-4a43-a9d5-25ab1be825b3","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"bridge","parameters":""},{"uuid":"659be95d-dcf1-4c9b-b1fb-9d1f89ecc1bd","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"br_netfilter","parameters":""},{"uuid":"7f337f62-0b32-4552-86f3-754a3bcf409b","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"e1000","parameters":""},{"uuid":"ced07177-6ee2-4b23-af77-3a9a0b7a1708","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"e1000e","parameters":""},{"uuid":"f5a0c66d-431d-46cd-8ec8-0c88b1f4c9a7","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"forcedeth","parameters":""},{"uuid":"4141a814-372e-4585-819a-363b3d5e8367","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"igb","parameters":""},{"uuid":"ab2730c1-7038-4c03-a100-eb1a01895d9c","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"igbvf","parameters":""},{"uuid":"b3fcafc6-22f4-48b0-9af2-23214e656780","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ixgbe","parameters":""},{"uuid":"15a5ace5-f9b5-4443-a198-5ce420df9134","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ixgbevf","parameters":""},{"uuid":"cb850e63-7913-4f2a-8c02-e74b9ac316d5","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"i40e","parameters":""},{"uuid":"c71f5282-ecb2-46d1-9add-0743decad13e","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"tg3","parameters":""},{"uuid":"e6e98f56-8cc4-47af-8c11-a7df9929a699","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hv_netvsc","parameters":""},{"uuid":"7a97bc4f-ca2e-4db0-9fc1-c32f08417d67","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hv_storvsc","parameters":""},{"uuid":"bf1b6484-db9a-4791-a693-35127d988343","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hv_utils","parameters":""},{"uuid":"381c512d-4ae1-400c-a652-e8921278f152","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hv_vmbus","parameters":""},{"uuid":"eaa8755d-8465-47a9-bb62-40fd90c34488","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"btrfs","parameters":""},{"uuid":"e0c8291c-9c41-4fdc-ac0f-12b6528636f4","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"jfs","parameters":""},{"uuid":"153f2b2d-464e-4695-9925-ae896b53d829","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"nfs","parameters":""},{"uuid":"e9598a4e-5535-468c-a96e-86879649fb71","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"nfsv3","parameters":""},{"uuid":"7a2f62aa-576d-4c54-8310-d5b98ff29aa8","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"nfsv4","parameters":""},{"uuid":"c39c5e95-b7e8-419d-a176-b1e49b07e6e3","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"reiserfs","parameters":""},{"uuid":"ddf8c0f6-244b-44df-b740-41f506a949c2","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"xfs","parameters":""},{"uuid":"a33d39dd-153b-4e23-8b2a-cf0f4fc096a1","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"udf","parameters":""},{"uuid":"626394b8-36bf-4657-8670-1f332e6a0f7e","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"usbhid","parameters":""},{"uuid":"a63a296c-7edd-4537-ab25-69064b36dec2","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"dm-persistent-data","parameters":""},{"uuid":"3ae634e3-98a3-4a1f-a638-1b525cecd592","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"dm-bio-prison","parameters":""},{"uuid":"cad36e9c-5169-4f4b-9860-eee5e0838ff6","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"dm-bufio","parameters":""},{"uuid":"d0881f2a-5303-4dd8-9ea7-bce4fb462fc1","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"dm-thin-pool","parameters":""},{"uuid":"80d01ad6-1333-4e30-b1c4-061b6dba41f0","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"nvme","parameters":""},{"uuid":"2b2f2948-e165-4028-b4ae-1dc2829b2812","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"isofs","parameters":""},{"uuid":"b7d54db8-3a74-4590-9293-fb65581628a1","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"bnxt_en","parameters":""},{"uuid":"33ee6df3-3280-424d-adbd-259b727546ed","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"bonding","parameters":""},{"uuid":"e98a9eca-0636-4f11-a524-e7936de13634","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"hpilo","parameters":""},{"uuid":"1113adad-ea03-4091-89e9-ff0ba75c8bde","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ipmi_si","parameters":""},{"uuid":"15423999-a0c2-4f1a-b471-a46ed7404917","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ipmi_ssif","parameters":""},{"uuid":"8d91243c-a219-4646-b7f7-0364416fe8a0","baseType":"KernelModule","childType":"","to_be_removed":false,"modified":true,"revision":"","name":"ipmi_devintf","parameters":""}],"enableSOL":false,"SOLPort":"ttyS1","SOLSpeed":"115200","SOLFlowControl":true,"notes":"","fspart":"00000000-0000-0000-0000-000000000000","bootfspart":"00000000-0000-0000-0000-000000000000","revisionID":0,"parentSoftwareImage":"00000000-0000-0000-0000-000000000000","revisionHistory":[]},0]}

### documentation for entity data

https://172.21.15.254:8081/api/search?type=entity&name=SoftwareImage

Software Image Entity

SoftwareImage
Entity
Parent chain
Entity
SoftwareImage

Can be instantiated
Yes

## Children

Field name Datatype Default
SoftwareImage parameters

baseType "SoftwareImage"
readonly
childType ""
readonly
name string
unique
path string
unique
format:^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$
originalImage UUID
fileOperationInProgress bool
kernelVersion string
kernelParameters string
kernelOutputConsole string "tty0"
creationTime time 0
read only
not cloneable
display:time
modules List KernelModule
enableSOL bool false
SOLPort string "ttyS1"
SOLSpeed "115200"|"57600"|"38400"|"19200"|"9600"|"4800"|"2400"|"1200" "115200"
SOLFlowControl bool true
notes string
fspart UUID
not cloneable
ref:FSPart
bootfspart UUID
not cloneable
ref:FSPart
revisionID int64 0
not cloneable
parentSoftwareImage UUID
not cloneable
ref:SoftwareImage
revisionHistory List SoftwareImageRevisionInfo
not cloneable
Entity parameters

uuid UUID
read only
not cloneable
to_be_removed bool
not cloneable
modified bool
not cloneable
revision string
not cloneable
