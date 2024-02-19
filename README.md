# Automated-DNS-Service
AWS service to read LB tags and create Route53 HostedZones.

Service will run as a Daemon.

It will scan for LB's that has a specific tag

```
tag-key: automated-dns
tag-Value: <fqdn>
```

Once it identifies the LB it will then check if there is an existing route53 entry for the fqdn.

## If the HostedZone exists already
  - check if the Alias exists
  - - if yes then we dont want to mess up the DNS and the service wont make any changes
  - - if no then we'll create an Alias and add a tag to identify later that we created the Alias

If the HostedZone does not exists then
- Create the HostedZone and tag it.
- Create the Alias