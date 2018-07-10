package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceReportingTask() *schema.Resource {
	return &schema.Resource{
		Create: ResourceReportingTaskCreate,
		Read:   ResourceReportingTaskRead,
		Update: ResourceReportingTaskUpdate,
		Delete: ResourceReportingTaskDelete,
		Exists: ResourceReportingTaskExists,

		Schema: map[string]*schema.Schema{
			"parent_group_id": SchemaParentGroupId(),
			"revision":        SchemaRevision(),
			"component": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parent_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"comments": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"scheduling_period": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "0 sec",
						},
						"scheduling_strategy": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "TIMER_DRIVEN",
						},
						"properties": {
										Type:     schema.TypeMap,
										Required: true,
						},
					},
				},
			},
		},
	}
}

func ResourceReportingTaskCreate(d *schema.ResourceData, meta interface{}) error {
	reportingTask := ReportingTask{}
	reportingTask.Revision.Version = 0

	err := ReportingTaskFromSchema(d, &reportingTask)
	if err != nil {
		return fmt.Errorf("Failed to parse Reporting Task schema")
	}
	parentGroupId := reportingTask.Component.ParentGroupId

	client := meta.(*Client)
	err = client.CreateReportingTask(&reportingTask)
	if err != nil {
		return fmt.Errorf("Failed to create Reporting Task")
	}

	d.SetId(reportingTask.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceReportingTaskRead(d, meta)
}

func ResourceReportingTaskRead(d *schema.ResourceData, meta interface{}) error {
	reportingTaskId := d.Id()

	client := meta.(*Client)
	reportingTask, err := client.GetReportingTask(reportingTaskId)
	if err != nil {
		return fmt.Errorf("Error retrieving Reporting Task: %s", reportingTaskId)
	}

	err = ReportingTaskToSchema(d, reportingTask)
	if err != nil {
		return fmt.Errorf("Failed to serialize Reporting Task: %s", reportingTaskId)
	}

	return nil
}

func ResourceReportingTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	reportingTaskId := d.Id()

	client := meta.(*Client)
	reportingTask, err := client.GetReportingTask(reportingTaskId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Reporting Task: %s", reportingTaskId)
		}
	}

	err = ReportingTaskFromSchema(d, reportingTask)
	if err != nil {
		return fmt.Errorf("Failed to parse Reporting Task schema: %s", reportingTaskId)
	}

	err = client.UpdateReportingTask(reportingTask)
	if err != nil {
		return fmt.Errorf("Failed to update Reporting Task: %s", reportingTaskId)
	}

	return ResourceReportingTaskRead(d, meta)
}

func ResourceReportingTaskDelete(d *schema.ResourceData, meta interface{}) error {
	reportingTaskId := d.Id()
	log.Printf("[INFO] Deleting Reporting Task: %s", reportingTaskId)

	client := meta.(*Client)
	reportingTask, err := client.GetReportingTask(reportingTaskId)
	if nil != err {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Reporting Task: %s", reportingTaskId)
		}
	}

	err = client.DeleteReportingTask(reportingTask)
	if err != nil {
		return fmt.Errorf("Error deleting Reporting Task: %s", reportingTaskId)
	}

	d.SetId("")
	return nil
}

func ResourceReportingTaskExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	reportingTaskId := d.Id()

	client := meta.(*Client)
	_, err := client.GetReportingTask(reportingTaskId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Reporting Task %s no longer exists, removing from state...", reportingTaskId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Reporting Task: %s", reportingTaskId)
		}
	}

	return true, nil
}

// Schema Helpers

func ReportingTaskFromSchema(d *schema.ResourceData, reportingTask *ReportingTask) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})

	parentGroupId := component["parent_group_id"].(string)
	reportingTask.Component.ParentGroupId = parentGroupId
	reportingTask.Component.Name = component["name"].(string)
    reportingTask.Component.Type = component["type"].(string)

	reportingTask.Component.Properties = map[string]interface{}{}
	properties := component["properties"].(map[string]interface{})
	for k, v := range properties {
		reportingTask.Component.Properties[k] = v.(string)
	}

	reportingTask.Component.SchedulingStrategy = component["scheduling_strategy"].(string)
	reportingTask.Component.SchedulingPeriod = component["scheduling_period"].(string)

	return nil
}

func ReportingTaskToSchema(d *schema.ResourceData, reportingTask *ReportingTask) error {
	revision := []map[string]interface{}{{
		"version": reportingTask.Revision.Version,
	}}
	d.Set("revision", revision)

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"name":            reportingTask.Component.Name,
		"type":            reportingTask.Component.Type,
		"properties":      reportingTask.Component.Properties,
		"scheduling_strategy": reportingTask.Component.SchedulingStrategy,
		"scheduling_period": reportingTask.Component.SchedulingPeriod,
	}}
	d.Set("component", component)

	return nil
}
