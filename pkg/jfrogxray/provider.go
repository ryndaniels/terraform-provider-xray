package jfrogxray

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xero-oss/go-xray/xray"

	"github.com/atlassian/go-artifactory/v2/artifactory/transport"
)

// Xray Provider that supports configuration via username+password or a token
// Supported resources are (for now) watches and policies
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("XRAY_URL", nil),
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("XRAY_USERNAME", nil),
				ConflictsWith: []string{"access_token"},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("XRAY_PASSWORD", nil),
				ConflictsWith: []string{"access_token"},
			},
			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				DefaultFunc:   schema.EnvDefaultFunc("XRAY_ACCESS_TOKEN", nil),
				ConflictsWith: []string{"username", "password"},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"xray_watch": resourceXrayWatch(),
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ConfigureFunc: providerConfigure,
	}
}

// Creates the client for xray, will prefer token auth over basic auth if both set
func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	if d.Get("url") == nil {
		return nil, fmt.Errorf("url cannot be nil")
	}

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	accessToken := d.Get("access_token").(string)

	var client *http.Client
	if username != "" && password != "" {
		tp := transport.BasicAuth{
			Username: username,
			Password: password,
		}
		client = tp.Client()
	} else if accessToken != "" {
		tp := transport.AccessTokenAuth{
			AccessToken: accessToken,
		}
		client = tp.Client()
	} else {
		return nil, fmt.Errorf("either [username, password] or [access_token] must be set to use provider")
	}

	rt, err := xray.NewClient(d.Get("url").(string), client)

	if err != nil {
		return nil, err
	} else if _, resp, err := rt.V1.System.Ping(context.Background()); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to ping server. Got %d", resp.StatusCode)
	}

	return rt, nil
}
