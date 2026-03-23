# Copyright (c) ZStack.io, Inc.

run "create_cluster_for_crud" {
  command = apply

  assert {
    condition     = zsphere_cluster.test.uuid != null
    error_message = "UUID should be computed after creation"
  }
}

run "read_cluster" {
  command = plan

  assert {
    condition     = zsphere_cluster.test.name == "test-cluster"
    error_message = "Cluster name should match"
  }

  assert {
    condition     = zsphere_cluster.test.state != null
    error_message = "Cluster state should be readable"
  }
}

run "update_cluster" {
  command = apply

  variables {
    cluster_name        = "crud-updated-cluster"
    cluster_description = "CRUD test description"
  }

  assert {
    condition     = zsphere_cluster.test.name == "crud-updated-cluster"
    error_message = "Cluster name should be updated through CRUD"
  }

  assert {
    condition     = zsphere_cluster.test.description == "CRUD test description"
    error_message = "Cluster description should be updated through CRUD"
  }
}
