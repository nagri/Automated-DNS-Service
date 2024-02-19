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
  - - if yes then we don't want to mess up the DNS and the service wont make any changes
  - - if no then we'll create an Alias and add a tag to identify later that we created the Alias

If the HostedZone does not exists then
- Create the HostedZone and tag it.
- Create the Alias

# How to Run 
go version used `go version go1.20.6 darwin/amd64`

set AWS Env variables

```
export AWS_ACCESS_KEY_ID=<AWS_ACCESS_KEY_ID>
export AWS_SECRET_ACCESS_KEY=<AWS_SECRET_ACCESS_KEY>
export AWS_DEFAULT_REGION=<AWS_DEFAULT_REGION>
```

change directory to where the `main.go` file resides.
then run with
```
go run main.go
```

This will spin up a log running process that scans the AWS LB every 1 minutes. If there exists an LB that has a tag `automated-dns` with its value set to `domain.extention`;

This process will create a HostedZone if it does not exist.
It will then create an Alias in the HostedZone and add few tags in the HostedZone to identify if its being managed by this Automated-DNS service. 

If you change the tag value on an existing LB, from `domain.extention` to `svc.domain.extention` the Automated-DNS service will pick up the change and update the HostedZone as well.

A lot of corner cases are not implemented.
1. If the LB is deleted then, the Automated-DNS service should identify it and delete the relevent entries/HostedZone from Route53
2. If the automated-dns Value is changed Altogether `domain.extention` to `anotherdomain.ext` then the service should delete the older hostedZone and recreate the new HostedZone.
3. Have not implemented test cases due to time crunch. 

