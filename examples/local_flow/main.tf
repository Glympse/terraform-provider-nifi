provider "nifi" {
  host = "${var.nifi_host}"
}

resource "nifi_process_group" "local_flow" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    name = "local_flow"

    position {
      x = 0
      y = 0
    }
  }
}

resource "nifi_processor" "generate_flowfile" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"
    name = "generate_flowfile"
    type = "org.apache.nifi.processors.standard.GenerateFlowFile"

    position {
      x = 0
      y = 0
    }

    config {
      properties {
        "File Size" = "0B"
        "Batch Size" = "1"
        "Data Format" = "Text"
        "Unique FlowFiles" = "false"
      }

      auto_terminated_relationships = [
        "success"
      ]
    }
  }
}

resource "nifi_processor" "merge_content_1" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"
    name = "merge_content_1"
    type = "org.apache.nifi.processors.standard.MergeContent"

    position {
      x = 0
      y = 250
    }

    config {
      properties {
        "Merge Strategy" = "Bin-Packing Algorithm"
        "Merge Format" = "Binary Concatenation"
        "Attribute Strategy" = "Keep Only Common Attributes"
        "Minimum Number of Entries" = "1"
        "Maximum Number of Entries" = "1000"
        "Minimum Group Size" = "0 B"
        "Maximum number of Bins" = "5"
        "Delimiter Strategy" = "Filename"
        "Compression Level" = "1"
        "Keep Path" = "false"
      }

      auto_terminated_relationships = [
        "failure"
      ]
    }
  }
}
