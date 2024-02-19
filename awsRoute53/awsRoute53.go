/*
If the HostedZone exists already
  - check if the Alias exists
  - - if yes then we dont want to mess up the DNS and the service wont make any changes
  - - if no then we'll create an Alias and add a tag to identify later that we created the Alias

If the HostedZone does not exists then
- Create the HostedZone and tag it.
- Create the Alias
*/
package awsroute53

// This will manage DNS on Route53
func ManageDNS() {

}

// Check if the Hosted zone exists
func checkIfHostedZoneExist() {

}

// checkIfAliasOrARecordExist
func checkIfAliasOrARecordExist() {

}

// To create Alias in Route53
func createAlias() {

}

// To create tags for the resouce
func createTags() {

}
