package jfrogxray

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xero-oss/go-xray/xray"
	"github.com/xero-oss/go-xray/xray/v2"
)

func resourceXrayWatch() *schema.Resource {
	return &schema.Resource{
		Create: resourceXrayWatchCreate,
		Read:   resourceXrayWatchRead,
		Update: resourceXrayWatchUpdate,
		Delete: resourceXrayWatchDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"resources": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bin_mgr_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"filters": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"assigned_policies": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func unpackWatch(d *schema.ResourceData) *v2.Watch {
	watch := new(v2.Watch)

	gd := &v2.WatchGeneralData{
		Name: xray.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		gd.Description = xray.String(v.(string))
	}
	if v, ok := d.GetOk("active"); ok {
		gd.Active = xray.Bool(v.(bool))
	}
	watch.GeneralData = gd

	pr := &v2.WatchProjectResources{}
	if v, ok := d.GetOk("resources"); ok {
		r := &[]v2.WatchProjectResource{}
		for _, res := range v.([]interface{}) {
			*r = append(*r, *unpackProjectResource(res))
		}
		pr.Resources = r
	}
	watch.ProjectResources = pr

	ap := &[]v2.WatchAssignedPolicy{}
	if v, ok := d.GetOk("assigned_policies"); ok {
		for _, pol := range v.([]interface{}) {
			*ap = append(*ap, *unpackAssignedPolicy(pol))
		}
	}
	watch.AssignedPolicies = ap

	return watch
}

func unpackProjectResource(rawCfg interface{}) *v2.WatchProjectResource {
	resource := new(v2.WatchProjectResource)

	cfg := rawCfg.(map[string]interface{})
	resource.Type = xray.String(cfg["type"].(string))
	if v, ok := cfg["bin_mgr_id"]; ok {
		resource.BinaryManagerId = xray.String(v.(string))
	}
	if v, ok := cfg["name"]; ok {
		resource.Name = xray.String(v.(string))
	}
	if v, ok := cfg["filters"]; ok {
		filters := &[]v2.WatchFilter{}
		for _, f := range v.([]interface{}) {
			*filters = append(*filters, *unpackFilter(f))
		}
		resource.Filters = filters
	}

	return resource
}

func unpackFilter(rawCfg interface{}) *v2.WatchFilter {
	filter := new(v2.WatchFilter)

	cfg := rawCfg.(map[string]interface{})
	filter.Type = xray.String(cfg["type"].(string))

	wf := new(v2.WatchFilterValueWrapper)
	if err := wf.UnmarshalJSON([]byte(cfg["value"].(string))); err == nil {
		filter.Value = wf
	}

	return filter
}

func unpackAssignedPolicy(rawCfg interface{}) *v2.WatchAssignedPolicy {
	policy := new(v2.WatchAssignedPolicy)

	cfg := rawCfg.(map[string]interface{})
	policy.Name = xray.String(cfg["name"].(string))
	policy.Type = xray.String(cfg["type"].(string))

	return policy
}

func packProjectResources(resources *v2.WatchProjectResources) []interface{} {
	if resources == nil || resources.Resources == nil {
		return []interface{}{}
	}

	packedResources := []interface{}{}
	for _, res := range *resources.Resources {
		m := make(map[string]interface{})
		m["type"] = res.Type
		if res.Name != nil {
			m["name"] = res.Name
		}
		if res.BinaryManagerId != nil {
			m["bin_mgr_id"] = res.BinaryManagerId
		}
		m["filters"] = packFilters(res.Filters)
		packedResources = append(packedResources, m)
	}

	return packedResources
}

func packFilters(filters *[]v2.WatchFilter) []interface{} {
	if filters == nil {
		return []interface{}{}
	}

	packedFilters := []interface{}{}
	for _, f := range *filters {
		m := make(map[string]interface{})
		m["type"] = f.Type
		m["value"] = f.Value.WatchFilterValue
		packedFilters = append(packedFilters, m)
	}

	return packedFilters
}

func packAssignedPolicies(policies *[]v2.WatchAssignedPolicy) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	packedPolicies := []interface{}{}
	for _, p := range *policies {
		m := make(map[string]interface{})
		m["name"] = p.Name
		m["type"] = p.Type
		packedPolicies = append(packedPolicies, m)
	}

	return packedPolicies
}

func resourceXrayWatchCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	watch := unpackWatch(d)

	_, err := c.V2.Watches.CreateWatch(context.Background(), watch)
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name) // ID may be returned according to the API docs, but not in go-xray
	return resourceXrayWatchRead(d, meta)
}

func resourceXrayWatchRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	watch, resp, err := c.V2.Watches.GetWatch(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Xray watch (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	if err := d.Set("description", watch.GeneralData.Description); err != nil {
		return err
	}
	if err := d.Set("active", watch.GeneralData.Active); err != nil {
		return err
	}
	if err := d.Set("resources", packProjectResources(watch.ProjectResources)); err != nil {
		return err
	}
	if err := d.Set("assigned_policies", packAssignedPolicies(watch.AssignedPolicies)); err != nil {
		return err
	}

	return nil
}

func resourceXrayWatchUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	watch := unpackWatch(d)
	_, err := c.V2.Watches.UpdateWatch(context.Background(), d.Id(), watch)
	if err != nil {
		return err
	}

	d.SetId(*watch.GeneralData.Name)
	return resourceXrayWatchRead(d, meta)
}

func resourceXrayWatchDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	resp, err := c.V2.Watches.DeleteWatch(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}
