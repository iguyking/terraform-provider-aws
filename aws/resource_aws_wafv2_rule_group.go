package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafv2RuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2RuleGroupCreate,
		Read:   resourceAwsWafv2RuleGroupRead,
		Update: resourceAwsWafv2RuleGroupUpdate,
		Delete: resourceAwsWafv2RuleGroupDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"block": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"count": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"statement": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"byte_match_statement": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"field_to_match": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"all_query_arguments": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
															},
														},
													},
												},
												"positional_constraint": {
													Type:     schema.TypeString,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
													ValidateFunc: validation.StringInSlice([]string{
														wafv2.PositionalConstraintContains,
														wafv2.PositionalConstraintContainsWord,
														wafv2.PositionalConstraintEndsWith,
														wafv2.PositionalConstraintExactly,
														wafv2.PositionalConstraintStartsWith,
													}, false),
												},
												"search_string": {
													Type:         schema.TypeString,
													Required:     true,
													Elem:         &schema.Schema{Type: schema.TypeString},
													ValidateFunc: validation.StringLenBetween(1, 50),
												},
												"text_transformation": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"priority": {
																Type:     schema.TypeInt,
																Required: true,
															},
															"type": {
																Type:     schema.TypeString,
																Required: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
																ValidateFunc: validation.StringInSlice([]string{
																	wafv2.TextTransformationTypeCmdLine,
																	wafv2.TextTransformationTypeCompressWhiteSpace,
																	wafv2.TextTransformationTypeHtmlEntityDecode,
																	wafv2.TextTransformationTypeLowercase,
																	wafv2.TextTransformationTypeNone,
																	wafv2.TextTransformationTypeUrlDecode,
																}, false),
															},
														},
													},
												},
											},
										},
									},
									"geo_match_statement": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"country_codes": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"visibility_config": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_metrics_enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"metric_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"sampled_requests_enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"visibility_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_metrics_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"metric_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"sampled_requests_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsWafv2RuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateRuleGroupOutput

	params := &wafv2.CreateRuleGroupInput{
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		Capacity:         aws.Int64(int64(d.Get("capacity").(int))),
		Rules:            expandWafv2Rules(d.Get("rule").([]interface{})),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateRuleGroup(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationException, "An error occurred during the tagging operation") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "AWS WAF couldn’t perform your tagging operation because of an internal error") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateRuleGroup(params)
	}
	if err != nil {
		return err
	}
	d.SetId(*resp.Summary.Id)

	return resourceAwsWafv2RuleGroupRead(d, meta)
}

func resourceAwsWafv2RuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	params := &wafv2.GetRuleGroupInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetRuleGroup(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
			log.Printf("[WARN] WAFV2 RuleGroup (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", resp.RuleGroup.Name)
	d.Set("capacity", resp.RuleGroup.Capacity)
	d.Set("description", resp.RuleGroup.Description)
	d.Set("arn", resp.RuleGroup.ARN)
	d.Set("rule", flattenWafv2Rules(resp.RuleGroup.Rules))
	d.Set("visibility_config", flattenWafv2VisibilityConfig(resp.RuleGroup.VisibilityConfig))

	tags, err := keyvaluetags.Wafv2ListTags(conn, *resp.RuleGroup.ARN)
	if err != nil {
		return fmt.Errorf("error listing tags for WAFV2 RuleGroup (%s): %s", *resp.RuleGroup.ARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWafv2RuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.GetRuleGroupOutput
	params := &wafv2.GetRuleGroupInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}
	log.Printf("[INFO] Updating WAFV2 RuleGroup %s", d.Id())

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.GetRuleGroup(params)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting lock token: %s", err))
		}

		u := &wafv2.UpdateRuleGroupInput{
			Id:               aws.String(d.Id()),
			Name:             aws.String(d.Get("name").(string)),
			Scope:            aws.String(d.Get("scope").(string)),
			LockToken:        resp.LockToken,
			VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
			Rules:            expandWafv2Rules(d.Get("rule").([]interface{})),
		}

		if v, ok := d.GetOk("description"); ok && len(v.(string)) > 0 {
			u.Description = aws.String(d.Get("description").(string))
		}

		_, err = conn.UpdateRuleGroup(u)

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.UpdateRuleGroup(&wafv2.UpdateRuleGroupInput{
			Id:               aws.String(d.Id()),
			Name:             aws.String(d.Get("name").(string)),
			Scope:            aws.String(d.Get("scope").(string)),
			Description:      aws.String(d.Get("description").(string)),
			LockToken:        resp.LockToken,
			Rules:            expandWafv2Rules(d.Get("rule").([]interface{})),
			VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
		})
	}

	if err != nil {
		return fmt.Errorf("Error updating WAFV2 RuleGroup: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafv2RuleGroupRead(d, meta)
}

func resourceAwsWafv2RuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.GetRuleGroupOutput
	params := &wafv2.GetRuleGroupInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}
	log.Printf("[INFO] Deleting WAFV2 RuleGroup %s", d.Id())

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.GetRuleGroup(params)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting lock token: %s", err))
		}

		_, err = conn.DeleteRuleGroup(&wafv2.DeleteRuleGroupInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRuleGroup(&wafv2.DeleteRuleGroupInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFV2 RuleGroup: %s", err)
	}

	return nil
}

func expandWafv2VisibilityConfig(l []interface{}) *wafv2.VisibilityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &wafv2.VisibilityConfig{}

	if v, ok := m["cloudwatch_metrics_enabled"]; ok {
		configuration.CloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["metric_name"]; ok && len(v.(string)) > 0 {
		configuration.MetricName = aws.String(v.(string))
	}

	if v, ok := m["sampled_requests_enabled"]; ok {
		configuration.SampledRequestsEnabled = aws.Bool(v.(bool))
	}

	return configuration
}

func expandWafv2Rules(l []interface{}) []*wafv2.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWafv2Rule(rule.(map[string]interface{})))
	}

	return rules
}

func expandWafv2Rule(m map[string]interface{}) *wafv2.Rule {
	if m == nil {
		return nil
	}

	return &wafv2.Rule{
		Name:             aws.String(m["name"].(string)),
		Priority:         aws.Int64(int64(m["priority"].(int))),
		Action:           expandWafv2RuleAction(m["action"].([]interface{})),
		Statement:        expandWafv2Statement(m["statement"].([]interface{})),
		VisibilityConfig: expandWafv2VisibilityConfig(m["visibility_config"].([]interface{})),
	}
}

func expandWafv2RuleAction(l []interface{}) *wafv2.RuleAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.RuleAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = &wafv2.AllowAction{}
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = &wafv2.BlockAction{}
	}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = &wafv2.CountAction{}
	}

	return action
}

func expandWafv2Statement(l []interface{}) *wafv2.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &wafv2.Statement{}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandWafv2ByteMatchStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandWafv2GeoMatchStatement(v.([]interface{}))
	}

	return statement
}

func expandWafv2ByteMatchStatement(l []interface{}) *wafv2.ByteMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.ByteMatchStatement{
		FieldToMatch:         expandWafv2FieldToMatch(m["field_to_match"].([]interface{})),
		PositionalConstraint: aws.String(m["positional_constraint"].(string)),
		SearchString:         []byte(m["search_string"].(string)),
		TextTransformations:  expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}
}

func expandWafv2FieldToMatch(l []interface{}) *wafv2.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	return &wafv2.FieldToMatch{
		AllQueryArguments: &wafv2.AllQueryArguments{},
	}
}

func expandWafv2TextTransformations(l []interface{}) []*wafv2.TextTransformation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.TextTransformation, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWafv2TextTransformation(rule.(map[string]interface{})))
	}

	return rules
}

func expandWafv2TextTransformation(m map[string]interface{}) *wafv2.TextTransformation {
	if m == nil {
		return nil
	}

	return &wafv2.TextTransformation{
		Priority: aws.Int64(int64(m["priority"].(int))),
		Type:     aws.String(m["type"].(string)),
	}
}

func expandWafv2GeoMatchStatement(l []interface{}) *wafv2.GeoMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.GeoMatchStatement{
		CountryCodes: expandStringList(m["country_codes"].([]interface{})),
	}
}

func flattenWafv2Rules(r []*wafv2.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["action"] = flattenWafv2RuleAction(rule.Action)
		m["name"] = aws.StringValue(rule.Name)
		m["priority"] = int(aws.Int64Value(rule.Priority))
		m["statement"] = flattenWafv2Statement(rule.Statement)
		m["visibility_config"] = flattenWafv2VisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenWafv2RuleAction(a *wafv2.RuleAction) interface{} {

	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = make([]map[string]interface{}, 1)
	}

	if a.Block != nil {
		m["block"] = make([]map[string]interface{}, 1)
	}

	if a.Count != nil {
		m["count"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenWafv2Statement(s *wafv2.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenWafv2ByteMatchStatement(s.ByteMatchStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenWafv2GeoMatchStatement(s.GeoMatchStatement)
	}

	return []interface{}{m}
}

func flattenWafv2ByteMatchStatement(b *wafv2.ByteMatchStatement) interface{} {
	if b == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":        flattenWafv2FieldToMatch(b.FieldToMatch),
		"positional_constraint": b.PositionalConstraint,
		"search_string":         string(b.SearchString),
		"text_transformation":   flattenWafv2TextTransformations(b.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2FieldToMatch(f *wafv2.FieldToMatch) interface{} {
	if f == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"all_query_arguments": make([]map[string]interface{}, 1),
	}

	return []interface{}{m}
}

func flattenWafv2TextTransformations(l []*wafv2.TextTransformation) []interface{} {
	out := make([]interface{}, len(l))
	for i, t := range l {
		m := make(map[string]interface{})
		m["priority"] = int(aws.Int64Value(t.Priority))
		m["type"] = aws.StringValue(t.Type)
		out[i] = m
	}
	return out
}

func flattenWafv2GeoMatchStatement(g *wafv2.GeoMatchStatement) interface{} {
	if g == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"country_codes": flattenStringList(g.CountryCodes),
	}

	return []interface{}{m}
}

func flattenWafv2VisibilityConfig(config *wafv2.VisibilityConfig) interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_metrics_enabled": aws.BoolValue(config.CloudWatchMetricsEnabled),
		"metric_name":                aws.String(*config.MetricName),
		"sampled_requests_enabled":   aws.BoolValue(config.SampledRequestsEnabled),
	}

	return []interface{}{m}
}
