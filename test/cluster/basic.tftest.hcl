# Copyright (c) ZStack.io, Inc.

run "create_basic_cluster" {
  command = apply

  assert {
    condition     = zsphere_cluster.test.name == "test-cluster"
    error_message = "Cluster name should be 'test-cluster'"
  }

  assert {
    condition     = zsphere_cluster.test.hypervisor_type == "KVM"
    error_message = "Cluster hypervisor type should be 'KVM'"
  }

  assert {
    condition     = zsphere_cluster.test.type == "zstack"
    error_message = "Cluster type should be 'zstack'"
  }

  assert {
    condition     = zsphere_cluster.test.architecture == "x86_64"
    error_message = "Cluster architecture should be 'x86_64'"
  }
}

run "verify_computed_attributes" {
  command = apply

  assert {
    condition     = zsphere_cluster.test.uuid != null && zsphere_cluster.test.uuid != ""
    error_message = "UUID should be computed and non-empty"
  }

  assert {
    condition     = zsphere_cluster.test.state != null
    error_message = "State should be computed"
  }

  assert {
    condition     = zsphere_cluster.test.datacenter_uuid != null && zsphere_cluster.test.datacenter_uuid != ""
    error_message = "Zone UUID should be set"
  }
}

run "update_cluster_name" {
  command = apply

  variables {
    cluster_name = "test-cluster-updated"
  }

  assert {
    condition     = zsphere_cluster.test.name == "test-cluster-updated"
    error_message = "Cluster name should be updated"
  }

  assert {
    condition     = zsphere_cluster.test.uuid != null
    error_message = "UUID should remain unchanged"
  }
}

run "update_cluster_description" {
  command = apply

  variables {
    cluster_description = "Updated description"
  }

  assert {
    condition     = zsphere_cluster.test.description == "Updated description"
    error_message = "Cluster description should be updated"
  }
}
