package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Description: "The user name for the Infoblox API",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_USER", nil),
			},
			"password": &schema.Schema{
				Description: "The user password for the Infoblox API",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_PASSWORD", nil),
			},
			"server": &schema.Schema{
				Description: "The Infoblox WAPI server",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_SERVER", nil),
			},
			"protocol": &schema.Schema{
				Description: "The protocol for communicating with Infoblox (https/http)",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_PROTOCOL", "https"),
			},
			"version": &schema.Schema{
				Description: "The Infoblox WAPI version",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_VERSION", "1.2.1"),
			},
			"allow_unverified_ssl": &schema.Schema{
				Description: "If set, permit unverifiable SSL certificates from Infoblox",
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_ALLOW_UNVERIFIED_SSL", false),
			},
			"timeout": &schema.Schema{
				Description: "The HTTP-Client timeout length in seconds for individual requests",
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_TIMEOUT", 30),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"infoblox_host_record": resourceInfobloxHostRecord(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	verString := d.Get("version").(string)
	ver := strings.Split(verString, ".")
	major, magErr := strconv.Atoi(ver[0])
	minor, minErr := strconv.Atoi(ver[1])
	patch, patErr := strconv.Atoi(ver[2])
	if magErr != nil || major < 1 || minErr != nil || (major == 1 && minor < 2) || patErr != nil || (major == 1 && minor == 2 && patch < 1) {
		return nil, fmt.Errorf("Unsupported Infoblox version %s. Infoblox WAPI 1.2.1 or higher is required", verString)
	}
	endpoint := d.Get("protocol").(string) + "://" + d.Get("server").(string) + "/wapi/v" + verString

	config := Config{
		User:             d.Get("user").(string),
		Password:         d.Get("password").(string),
		InfobloxEndpoint: endpoint,
		InfobloxVersion:  Semver{Major: major, Minor: minor, Patch: patch},
		InsecureFlag:     d.Get("allow_unverified_ssl").(bool),
		HTTPTimeout:      d.Get("timeout").(int),
	}
	return config.Client()
}
