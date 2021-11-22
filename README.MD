## Installation


```bash
git clone https://github.com/axxyhtrx/netbox-rollcall.git
cd netbox-rollcall
```

## Pre-requirements
- Create `config.yaml` file in a root of the project. Configure IP address/FQDN (netboxhost) of Netbox instance and API key(netboxapitoken).

- Create `VRFs` in a Netbox, and use the same names in a config file with same names.
- Configure `scanthreads` value for NMAP, to limit/increase threading for Nmap scanner.

```yaml
netbox:
  netboxhost: '10.80.37.00'
  netboxapitoken: '12014TOKENTOKENff4c'

targets:
  - vrf: ANTWERP
    subnets:
      - 192.168.1.1/32
      - 20.3.4.0/24
  - vrf: MOSCOW
    subnets:
      - 192.168.2.1/32
      - 21.3.4.0/24


scanthreads: 20

```


- Create tag with the name `Scanned` in a Netbox.

- [NMAP](Nmap.org) should be installed.


## Usage
```bash
go run main.go scan
```
or compile with go build main.go and use as:

```
netbox-rollcall scan
```
## Contributing

PRs are welcome!