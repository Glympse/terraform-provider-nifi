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

      auto_terminated_relationships = []
    }
  }
}

resource "nifi_connection" "generate_to_merge_1" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.generate_flowfile.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content_1.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    selected_relationships = [
      "success"
    ]
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
        "failure", "merged"
      ]
    }
  }
}

resource "nifi_connection" "merge_1_to_merge_2" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content_1.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content_2.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    selected_relationships = [
      "original"
    ]
  }
}

resource "nifi_processor" "merge_content_2" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"
    name = "merge_content_2"
    type = "org.apache.nifi.processors.standard.MergeContent"

    position {
      x = 0
      y = 500
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
        "merged", "original"
      ]
    }
  }
}

resource "nifi_connection" "merge_2_to_merge_2" {
  component {
    parent_group_id = "${nifi_process_group.local_flow.id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content_2.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content_2.id}"
      group_id = "${nifi_process_group.local_flow.id}"
    }

    selected_relationships = [
      "failure"
    ]

    bends = [{
      x = 460
      y = 540
    }, {
      x = 460
      y = 600
    }]
  }
}
