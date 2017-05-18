# terraform-infoblox
This is a Terraform Provider for creating Host Records with the Infoblox REST API.  
(minimum wapi v1.2.1)

#### design
This is a very simple provider that leverages the Infoblox REST API directly.  
It does not use a vendored/external API for abstracting the transport.

It currently only supports creating host records (A/PTR collections in Infoblox).  
Contributions are welcome!

Next steps are to add auto-discovery of valid hostnames and IP addresses using the infoblox search API's

#### build
The `./build` script uses docker to cache deps in `./gosrc_amd64` and build the code.  
The resulting artifact will be at `./sandbox/terraform-provider-infoblox` where it can be used with the example `infoblox.tf`.

#### runtime
Dependencies on [resty](https://github.com/go-resty/resty) result in dynamic bindings to net in glibc. (guessing)  
This will cause Terraform to fail to exec the provider in alpine containers like `hashicorp/terraform`.

Use `stealthybox/infra` for an alpine terraform with glibc:
```bash
docker run -v/$PWD://terra -w//terra stealthybox/infra terraform plan
```
... or just run terraform on your local machine like a normal person.
