#Tip

##How to ignore poisition changes in application

```
 resource "nifi_resource" "name" {
   component {
     name = "test_input_1"
     parent_group_id = "${var.nifi_root_process_group_id}"

     position { x=0 y=0 }
   }
   lifecycle {
     ignore_changes = ["component.0.position.0.x","component.0.position.0.y"]
   }
 }
```