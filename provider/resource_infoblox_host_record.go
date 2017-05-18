package main

import (
	"log"
	"strings"

	"fmt"

	"github.com/go-resty/resty"
	"github.com/hashicorp/terraform/helper/schema"
)

type Object struct {
	Ref string `json:"_ref"`
}
type Host struct {
	Object
	Name, View string
	Ttl        int
	Use_Ttl    bool
	Ipv4addrs  []Ipv4
}
type Ipv4 struct {
	Object
	Host, Ipv4addr     string
	Configure_for_dhcp bool
}
type WapiError struct {
	Error, Code, Text string
}

func resourceInfobloxHostRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceInfobloxHostRecordCreate,
		Read:   resourceInfobloxHostRecordRead,
		Update: resourceInfobloxHostRecordUpdate,
		Delete: resourceInfobloxHostRecordDelete,

		Schema: map[string]*schema.Schema{
			"domain": &schema.Schema{
				Description: "The domain name to create these records in",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": &schema.Schema{
				Description: "The subdomain of the record",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"ipv4": &schema.Schema{
				Description: "The ip-address or function used to generate one",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ttl": &schema.Schema{
				Description: "The TTL of the DNS record in seconds, used for client-cache invalidation",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     600,
			},
		},
	}
}

func resourceInfobloxHostRecordCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("\n[infoblox-provider] %s", "----------------- host record create")
	name := d.Get("name").(string) + "." + d.Get("domain").(string)
	ipv4 := d.Get("ipv4").(string)
	ttl := d.Get("ttl").(int)

	wapiErr := WapiError{}
	resp, err := resty.R().
		SetError(&wapiErr).
		SetBody(map[string]interface{}{
			"name": name,
			"ipv4addrs": []map[string]interface{}{
				map[string]interface{}{
					"ipv4addr": ipv4,
				},
			},
			"ttl":     ttl,
			"use_ttl": true,
		}).
		Post("/record:host")
	if handler := handleError(err, resp, wapiErr); handler != nil {
		return handler
	}
	d.SetId(strings.Replace(resp.String(), "\"", "", 2))
	return resourceInfobloxHostRecordRead(d, meta)
}

func resourceInfobloxHostRecordRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("\n[infoblox-provider] %s", "----------------- host record read")
	host := Host{}
	wapiErr := WapiError{}
	resp, err := resty.R().
		SetResult(&host).
		SetError(&wapiErr).
		SetQueryParams(map[string]string{
			"_return_fields+": "ttl,use_ttl",
		}).
		Get("/" + d.Id())
	log.Printf("\n[infoblox-provider] Wapi Object: %+v", host)
	if handler := handleError(err, resp, wapiErr); handler != nil {
		return handler
	}
	splitFqdn := strings.Split(host.Name, ".")
	d.Set("fqdn", host.Name)
	d.Set("name", splitFqdn[0])
	d.Set("domain", strings.Join(splitFqdn[1:], "."))
	d.Set("ipv4", host.Ipv4addrs[0].Ipv4addr)
	d.Set("ttl", host.Ttl)
	d.Set("view", host.View)
	return nil
}

func resourceInfobloxHostRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("\n[infoblox-provider] %s", "----------------- host record update")
	ipv4 := d.Get("ipv4").(string)

	wapiErr := WapiError{}
	resp, err := resty.R().
		SetError(&wapiErr).
		SetBody(map[string]interface{}{
			"ipv4addrs": []map[string]interface{}{
				map[string]interface{}{
					"ipv4addr": ipv4,
				},
			},
		}).
		Put("/" + d.Id())
	if handler := handleError(err, resp, wapiErr); handler != nil {
		return handler
	}
	return resourceInfobloxHostRecordRead(d, meta)
}

func resourceInfobloxHostRecordDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("\n[infoblox-provider] %s", "----------------- host record delete")
	wapiErr := WapiError{}
	resp, err := resty.R().
		SetError(&wapiErr).
		Delete("/" + d.Id())
	if handler := handleError(err, resp, wapiErr); handler != nil {
		return handler
	}
	return nil
}

func handleError(err error, resp *resty.Response, wapiErr WapiError) error {
	log.Printf("\n[infoblox-provider] HTTP Code: (%v) Response Body: %v", resp.StatusCode(), resp)
	if err != nil {
		return fmt.Errorf("[infoblox-provider] Resty Error: %+v", err)
	} else if resp.StatusCode() >= 300 && resp.StatusCode() < 400 {
		return fmt.Errorf("[infoblox-provider] HTTP Redirect: (%v)", resp.StatusCode())
	} else if resp.StatusCode() >= 400 && wapiErr.Error != "" {
		return fmt.Errorf("[infoblox-provider] WAPI Error: (%v) %+v", resp.StatusCode(), wapiErr)
	} else if resp.StatusCode() >= 400 {
		return fmt.Errorf("[infoblox-provider] Unknown HTTP Error: (%v) %+v", resp.StatusCode(), resp.String())
	}
	return nil
}
