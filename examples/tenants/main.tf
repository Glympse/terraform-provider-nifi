provider "nifi" {
  host = "127.0.0.1:9443"
  admin_cert = "/opt/nifi-toolkit/target/nifi-admin.pem"
  admin_key = "/opt/nifi-toolkit/target/nifi-admin.key"
  api_path =       "nifi-api"
}


resource "nifi_user" "test_user" {
  component {
    identity="test_user"
    parent_group_id = ""
    position {
      x = 0
      y = 0
    }
  }
}

resource "nifi_user" "test_user2" {
  component {
    identity="test_user2"
    parent_group_id = ""
    position {
      x = 0
      y = 0
    }
  }
}

resource "nifi_group" "test_group" {
  component {
    identity="test_group"
    parent_group_id = "root"
    position {
      x = 0
      y = 0
    }
    users=["${list(nifi_user.test_user.id, nifi_user.test_user2.id)}"]
  }
}

resource "nifi_group" "test_group2" {
  component {
    identity="test_group2"
    parent_group_id = "root"
    position {
      x = 0
      y = 0
    }
    users=["${list(nifi_user.test_user.id, nifi_user.test_user2.id)}"]
  }
}

resource "nifi_port" "test_input_1" {
  component {
    name = "test_input_1"
    parent_group_id = "${var.nifi_root_process_group_id}"
    type = "INPUT_PORT"
    position{ x=0 y=0}
  }
  lifecycle {
    ignore_changes = ["component.0.position.0.x","component.0.position.0.y", "component.0.type"]
  }
}

resource "nifi_port" "test_output_1" {
  component {
    name = "test_output_1"
    parent_group_id = "${var.nifi_root_process_group_id}"
    type = "OUTPUT_PORT"
    position{ x=0 y=0}
  }
  lifecycle {
    ignore_changes = ["component.0.position.0.x","component.0.position.0.y", "component.0.type"]
  }
}

resource "nifi_connection" "test_input_output" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    destination {
      type = "OUTPUT_PORT"
      id = "${nifi_port.test_output_1.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
    source {
      type = "INPUT_PORT"
      id = "${nifi_port.test_input_1.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
  }
}


resource "nifi_port" "test_input_2" {
  component {
    name = "test_input_2"
    parent_group_id = "${var.nifi_root_process_group_id}"
    type = "INPUT_PORT"
    position{ x=0 y=0}
  }
  lifecycle {
    ignore_changes = ["component.0.position.0.x","component.0.position.0.y", "component.0.type"]
  }
}

resource "nifi_port" "test_input_3" {
  component {
    name = "test_input_3"
    parent_group_id = "${var.nifi_root_process_group_id}"
    type = "INPUT_PORT"
    position{ x=0 y=0}
  }
  lifecycle {
    ignore_changes = ["component.0.position.0.x","component.0.position.0.y", "component.0.type"]
  }
}
resource "nifi_funnel" "test_funnel_1" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    position { x=0 y=0 }
  }
  lifecycle {
    ignore_changes = ["component.0.position.0.x","component.0.position.0.y"]
  }
}

resource "nifi_connection" "test_input_2_funnel_1" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    source {
      type = "INPUT_PORT"
      id = "${nifi_port.test_input_2.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
    destination {
      type = "FUNNEL"
      id = "${nifi_funnel.test_funnel_1.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
  }
}

resource "nifi_connection" "test_input_3_funnel_1" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    source {
      type = "INPUT_PORT"
      id = "${nifi_port.test_input_3.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
    destination {
      type = "FUNNEL"
      id = "${nifi_funnel.test_funnel_1.id}"
      group_id  = "${var.nifi_root_process_group_id}"
    }
  }
}
