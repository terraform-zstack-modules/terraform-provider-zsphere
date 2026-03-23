# Copyright (c) ZStack.io, Inc.

run "create_basic_cluster" {
  command = apply

  variables {
    cluster_name        = "test-basic-cluster"
    cluster_hypervisor_type = "KVM"
  }

  assert {
    condition     = zsphere_cluster.test.name == "test-basic-cluster"
    error_message = "Cluster name should be 'test-basic-cluster'"
  }

  assert {
    condition     = zsphere_cluster.test.hypervisor_type == "KVM"
    error_message = "Cluster hypervisor type should be 'KVM'"
  }
}

run "create_cluster_with_all_fields" {
  command = apply

  variables {
    cluster_name                        = "full-config-cluster"
    cluster_description                 = "Cluster with all fields configured"
    cluster_hypervisor_type             = "KVM"
    cluster_type                        = "zstack"
    cluster_architecture               = "x86_64"
    cluster_display_network_cidr        = "192.168.1.0/24"
    cluster_migrate_network_cidr       = "192.168.2.0/24"
    cluster_check_cpu_model            = "false"
    cluster_cpu_mode                   = "hostPassthrough"
    cluster_network_hp                 = false
    cluster_automation_level           = "closed"
    cluster_cpu_over_provisioning_ratio = "4"
    cluster_vm_ha_level               = "NeverStop"
    cluster_drs_enabled               = false
  }

  assert {
    condition     = zsphere_cluster.test.name == "full-config-cluster"
    error_message = "Cluster name should match"
  }

  assert {
    condition     = zsphere_cluster.test.hypervisor_type == "KVM"
    error_message = "Cluster hypervisor type should be KVM"
  }
}

run "create_cluster_with_drs" {
  command = apply

  variables {
    cluster_name              = "drs-cluster"
    cluster_hypervisor_type  = "KVM"
    cluster_drs_enabled      = true
    cluster_automation_level = "Automatic"
    cluster_drs_threshold_cpu    = 70
    cluster_drs_threshold_memory = 80
    cluster_drs_threshold_duration = 600
  }

  assert {
    condition     = zsphere_cluster.test.name == "drs-cluster"
    error_message = "Cluster name should match"
  }

  assert {
    condition     = zsphere_cluster.test.hypervisor_type == "KVM"
    error_message = "Cluster hypervisor type should be KVM"
  }
}
