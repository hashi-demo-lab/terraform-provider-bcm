# Tasks: Fix disksetup/raidconf XML Schema Tests

## Task Overview

| ID | Task | Status | Dependencies |
|----|------|--------|--------------|
| T1 | Update PartitionConfig test helper | Pending | None |
| T2 | Update PartitionConfigUpdated test helper | Pending | T1 |
| T3 | Update DiskSetupAdvanced test helper | Pending | T1 |
| T4 | Remove t.Skip() from PartitionConfiguration test | Pending | T1, T2 |
| T5 | Remove t.Skip() from DiskSetupAdvanced test | Pending | T3 |
| T6 | Run PartitionConfiguration acceptance test | Pending | T4 |
| T7 | Run DiskSetupAdvanced acceptance test | Pending | T5 |
| T8 | Update CLAUDE.md with XML schema docs | Pending | T6, T7 |
| T9 | Run full test suite validation | Pending | T6, T7 |
| T10 | Generate documentation with make generate | Pending | T9 |

---

## Task Details

### T1: Update PartitionConfig test helper

**File:** `internal/provider/resource_cmdevice_category_test.go`

**Function:** `testAccCMDeviceCategoryResourceConfig_PartitionConfig`

**Changes:**
Replace incorrect XML with valid BCM schema:

```hcl
disksetup = <<-EOT
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <blockdev>/dev/vda</blockdev>
    <partition id="a0" partitiontype="esp">
      <size>100M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="a1">
      <size>20G</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="a2">
      <size>max</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/local</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
  </device>
</diskSetup>
EOT
```

Also change `raidconf = "raid1"` to `raidconf = ""`

**Validation:** `go build ./internal/provider/`

---

### T2: Update PartitionConfigUpdated test helper

**File:** `internal/provider/resource_cmdevice_category_test.go`

**Function:** `testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated`

**Changes:**
Use valid XML with different partition layout for update testing:

```hcl
disksetup = <<-EOT
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <blockdev>/dev/vda</blockdev>
    <partition id="b0" partitiontype="esp">
      <size>200M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="b1">
      <size>30G</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="b2">
      <size>8G</size>
      <type>linux swap</type>
    </partition>
    <partition id="b3">
      <size>max</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/data</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
  </device>
</diskSetup>
EOT
```

Change `raidconf = "raid5"` to `raidconf = ""`

**Validation:** `go build ./internal/provider/`

---

### T3: Update DiskSetupAdvanced test helper

**File:** `internal/provider/resource_cmdevice_category_test.go`

**Function:** `testAccCMDeviceCategoryResourceConfig_DiskSetupAdvanced`

**Changes:**
Update XML format similar to T1, keeping advanced fields like `install_boot_record = true`

**Validation:** `go build ./internal/provider/`

---

### T4: Remove t.Skip() from PartitionConfiguration test

**File:** `internal/provider/resource_cmdevice_category_test.go`

**Function:** `TestAccCMDeviceCategory_PartitionConfiguration`

**Line:** ~837

**Change:**
```go
// REMOVE THIS LINE:
t.Skip("BLOCKED: Requires BCM XSD files for disksetup/raidconf validation. See issue #48")
```

Also update test state checks to remove raidconf assertions since we're using empty string:
- Remove `knownvalue.StringExact("raid1")` check
- Remove `knownvalue.StringExact("raid5")` check from update step

**Validation:** `go build ./internal/provider/`

---

### T5: Remove t.Skip() from DiskSetupAdvanced test

**File:** `internal/provider/resource_cmdevice_category_test.go`

**Function:** `TestAccCMDeviceCategoryResource_DiskSetupAdvanced`

**Line:** ~1110 (approximately)

**Change:**
```go
// REMOVE THIS LINE:
t.Skip("BLOCKED: Requires BCM XSD files for disksetup/raidconf validation. See issue #48")
```

Update assertions similarly to T4.

**Validation:** `go build ./internal/provider/`

---

### T6: Run PartitionConfiguration acceptance test

**Command:**
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 30m ./internal/provider/ \
-run "TestAccCMDeviceCategory_PartitionConfiguration"
```

**Expected:** PASS

**If fail:** Check error message, adjust XML format if needed

---

### T7: Run DiskSetupAdvanced acceptance test

**Command:**
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 30m ./internal/provider/ \
-run "TestAccCMDeviceCategoryResource_DiskSetupAdvanced"
```

**Expected:** PASS

**If fail:** Check error message, adjust XML format if needed

---

### T8: Update CLAUDE.md with XML schema docs

**File:** `/workspace/CLAUDE.md`

**Section:** Add after "### Known Patterns"

**Content:**
```markdown
### disksetup XML Schema

BCM validates `disksetup` against an internal XSD schema. Key format requirements:

**Root element:** `<diskSetup>` (camelCase, NOT `<disksetup>`)

**Structure:**
- `<device>` - Container for block devices and partitions
  - `<blockdev>` - Block device path (e.g., `/dev/sda`)
  - `<partition>` - Partition definition with required `id` attribute

**Partition children:**
- `<size>` - Size value: "100M", "20G", "max"
- `<type>` - Partition type: "linux", "linux swap"
- `<filesystem>` - FS type: "fat", "xfs", "ext4"
- `<mountPoint>` - Mount path: "/", "/boot/efi"
- `<mountOptions>` - Mount options: "defaults,noatime"

**Example:**
\```xml
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <partition id="a0" partitiontype="esp">
      <size>100M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
    </partition>
    <partition id="a1">
      <size>max</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/</mountPoint>
    </partition>
  </device>
</diskSetup>
\```

**raidconf field:** Currently uses empty string. XML format requires further BCM documentation.
```

---

### T9: Run full test suite validation

**Command:**
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/ \
-run "TestAccCMDeviceCategory"
```

**Expected:** All category tests PASS

---

### T10: Generate documentation with make generate

**Command:**
```bash
make generate
```

**Validation:**
- No errors
- Check `docs/` for updates
- Run `make fmt` and `make lint`

---

## Completion Checklist

- [ ] T1: PartitionConfig helper updated
- [ ] T2: PartitionConfigUpdated helper updated
- [ ] T3: DiskSetupAdvanced helper updated
- [ ] T4: t.Skip() removed from PartitionConfiguration
- [ ] T5: t.Skip() removed from DiskSetupAdvanced
- [ ] T6: PartitionConfiguration test passes
- [ ] T7: DiskSetupAdvanced test passes
- [ ] T8: CLAUDE.md updated
- [ ] T9: Full category test suite passes
- [ ] T10: Documentation generated

## Notes

- Tasks T1-T5 can be done together in a single edit session
- Tasks T6-T7 should be run sequentially to isolate failures
- If T6 or T7 fail, review BCM API error messages carefully
- The `raidconf` field uses empty string - create follow-up issue if RAID testing needed
