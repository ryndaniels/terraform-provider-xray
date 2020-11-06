package jfrogxray

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xero-oss/go-xray/xray"
	"github.com/xero-oss/go-xray/xray/v1"
)

func resourceXrayPolicy() *schema.Resource {
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
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"author": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"criteria": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Security criteria
									"min_severity": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"allow_unknown", "banned_licenses", "allowed_licenses"},
									},
									"cvss_range": {
										Type:          schema.TypeSet,
										Optional:      true,
										ConflictsWith: []string{"allow_unknown", "banned_licenses", "allowed_licenses"},
										MaxItems:      1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"to": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"from": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									// License Criteria
									"allow_unknown": {
										Type:          schema.TypeBool,
										Optional:      true,
										ConflictsWith: []string{"min_severity", "cvss_range"},
									},
									"banned_licenses": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"min_severity", "cvss_range"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"allowed_licenses": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"min_severity", "cvss_range"},
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"actions": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mails": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"fail_build": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"block_download": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unscanned": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"active": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"webhooks": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"custom_severity": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func unpackPolicy(d *schema.ResourceData) *v1.Policy {
	policy := new(v1.Policy)

	policy.Name = xray.String(d.Get("name").(string))
	if v, ok := d.GetOk("type"); ok {
		policy.Type = xray.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		policy.Description = xray.String(v.(string))
	}
	if v, ok := d.GetOk("author"); ok {
		policy.Author = xray.String(v.(string))
	}

	rules := &[]v1.PolicyRule{}
	for _, r := range d.Get("rules").([]interface{}) {
		*rules = append(*rules, *unpackRule(r))
	}
	policy.Rules = rules

	return policy
}

func unpackRule(rawCfg interface{}) *v1.PolicyRule {
	rule := new(v1.PolicyRule)

	cfg := rawCfg.(map[string]interface{})
	rule.Name = xray.String(cfg["name"].(string))
	rule.Priority = xray.Int(cfg["priority"].(int))

	rule.Criteria = unpackCriteria(cfg["criteria"].([]interface{}))
	if v, ok := cfg["actions"]; ok {
		rule.Actions = unpackActions(v.([]interface{}))
	}

	return rule
}

func unpackCriteria(rawCfg []interface{}) *v1.PolicyRuleCriteria {
	criteria := new(v1.PolicyRuleCriteria)

	cfg := rawCfg[0].(map[string]interface{}) // We made this a list of one to make schema validation easier

	if v, ok := cfg["min_severity"]; ok {
		criteria.MinimumSeverity = xray.String(v.(string))
	}
	if v, ok := cfg["cvss_range"]; ok {
		vMap := v.(map[string]interface{})
		criteria.CVSSRange = &v1.PolicyCVSSRange{
			To:   xray.Int(vMap["to"].(int)),
			From: xray.Int(vMap["from"].(int)),
		}
	}
	if v, ok := cfg["allow_unknown"]; ok {
		criteria.AllowUnkown = xray.Bool(v.(bool)) // "Unkown" is a typo in xray-oss
	}
	if v, ok := cfg["banned_licenses"]; ok {
		*criteria.BannedLicenses = v.([]string)
	}
	if v, ok := cfg["allowed_licenses"]; ok {
		*criteria.AllowedLicenses = v.([]string)
	}

	return criteria
}

func unpackActions(rawCfg []interface{}) *v1.PolicyRuleActions {
	actions := new(v1.PolicyRuleActions)

	cfg := rawCfg[0].(map[string]interface{}) // We made this a list of one to make schema validation easier

	if v, ok := cfg["mails"]; ok {
		*actions.Mails = v.([]string)
	}
	if v, ok := cfg["fail_build"]; ok {
		actions.FailBuild = xray.Bool(v.(bool))
	}
	if v, ok := cfg["block_download"]; ok {
		vMap := v.(map[string]interface{})
		actions.BlockDownload = &v1.BlockDownloadSettings{
			Unscanned: xray.Bool(vMap["unscanned"].(bool)),
			Active:    xray.Bool(vMap["active"].(bool)),
		}
	}
	if v, ok := cfg["webhooks"]; ok {
		*actions.Webhooks = v.([]string)
	}
	if v, ok := cfg["custom_severity"]; ok {
		actions.CustomSeverity = xray.String(v.(string))
	}

	return actions
}

func packRules(rules *[]v1.PolicyRule) []interface{} {
	if rules == nil {
		return []interface{}{}
	}

	packedRules := []interface{}{}
	for _, rule := range *rules {
		m := make(map[string]interface{})
		m["name"] = rule.Name
		m["priority"] = rule.Priority
		m["criteria"] = packCriteria(rule.Criteria)
		m["actions"] = packActions(rule.Actions)
		packedRules = append(packedRules, m)
	}

	return packedRules
}

func packCriteria(criteria *v1.PolicyRuleCriteria) []interface{} {
	if criteria == nil {
		return []interface{}{}
	}

	packedCriteria := []interface{}{}
	m := make(map[string]interface{})
	if criteria.MinimumSeverity != nil {
		m["min_severity"] = criteria.MinimumSeverity
	}
	if criteria.CVSSRange != nil {
		r := make(map[string]interface{})
		r["to"] = criteria.CVSSRange.To
		r["from"] = criteria.CVSSRange.From
		m["cvss_range"] = r
	}

	if criteria.AllowUnkown != nil { // Still a typo in the library
		m["allow_unknown"] = criteria.AllowUnkown
	}
	if criteria.BannedLicenses != nil {
		m["banned_licenses"] = criteria.BannedLicenses
	}
	if criteria.AllowedLicenses != nil {
		m["allowed_licenses"] = criteria.AllowedLicenses
	}

	packedCriteria = append(packedCriteria, m) // There's just one, but the schema type is a list
	return packedCriteria
}

func packActions(actions *v1.PolicyRuleActions) []interface{} {
	if actions == nil {
		return []interface{}{}
	}

	packedActions := []interface{}{}
	m := make(map[string]interface{})
	if actions.Mails != nil {
		m["mails"] = actions.Mails
	}
	if actions.FailBuild != nil {
		m["fail_build"] = actions.FailBuild
	}
	if actions.BlockDownload != nil {
		bd := make(map[string]interface{})
		bd["unscanned"] = actions.BlockDownload.Unscanned
		bd["active"] = actions.BlockDownload.Active
		m["block_download"] = bd
	}
	if actions.Webhooks != nil {
		m["webhooks"] = actions.Webhooks
	}
	if actions.CustomSeverity != nil {
		m["custom_severity"] = actions.CustomSeverity
	}

	packedActions = append(packedActions, m) // Another 1-item list in the schema
	return packedActions
}

func resourceXrayPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	policy := unpackPolicy(d)

	_, err := c.V1.Policies.CreatePolicy(context.Background(), policy)
	if err != nil {
		return err
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, meta)
}

func resourceXrayPolicyRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	policy, resp, err := c.V1.Policies.GetPolicy(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[WARN] Xray policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	if err := d.Set("type", policy.Type); err != nil {
		return err
	}
	if err := d.Set("description", policy.Description); err != nil {
		return err
	}
	if err := d.Set("author", policy.Author); err != nil {
		return err
	}
	if err := d.Set("created", policy.Created); err != nil {
		return err
	}
	if err := d.Set("modified", policy.Modified); err != nil {
		return err
	}
	if err := d.Set("rules", packRules(policy.Rules)); err != nil {
		return err
	}

	return nil
}

func resourceXrayPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	policy := unpackPolicy(d)
	_, err := c.V1.Policies.UpdatePolicy(context.Background(), d.Id(), policy)
	if err != nil {
		return err
	}

	d.SetId(*policy.Name)
	return resourceXrayPolicyRead(d, meta)
}

func resourceXrayPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*xray.Xray)

	resp, err := c.V1.Policies.DeletePolicy(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}
