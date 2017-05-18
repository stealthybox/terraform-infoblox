provider "infoblox" {
  user     = "user"
  password = "password"
  server   = "localhost"
  # version  = "1.4.1"  # default is 1.2.1
  # protocol = "http"   # default is https
  # timeout  = 5        # default is 30
  # allow_unverified_ssl = true  # default is false
} 

resource "infoblox_host_record" "mydomain" {
  count  = 3
  name   = "stealthybox-is-cool-${count.index+1}"
  domain = "mydomain.com"
  ipv4   = "10.0.0.${count.index+1}"
  ttl    = 600
}
