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
