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

import (
	awslb "Automated-DNS-Service/awsLB"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	dnsServiceManagedTagValue = "DNS-Service"
	dnsServiceManagedTagKey   = "Managed-By"
	aliasKey                  = "Alias"
	hostedZoneDescription     = "Managed via Automated DNS-Service"
	aliasExistMsg             = " Alias exists, wont make any entries."
	aRecordExist              = " A record exists, wont make any entries."
	recordTag                 = "recordManagedByDNSAutomation"
)

// This will manage DNS on Route53
func ManageDNS(hostedZone string, LbToManage *awslb.AWSLb) error {
	aliasExist := false
	hostedZoneID, err := checkIfHostedZoneExist(hostedZone)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if hostedZoneID == "" {
		createHostedZone(hostedZone, LbToManage)
		return nil
	} else {
		aliasExist = checkIfAliasOrARecordExist(hostedZoneID, LbToManage)
	}
	if aliasExist {
		return nil
	}
	createAlias(hostedZoneID, LbToManage)
	createRecordOnlyTags(hostedZoneID, LbToManage)
	return nil
}

// Check if the Hosted zone exists
func checkIfHostedZoneExist(hostedZone string) (string, error) {
	ctx := context.Background()
	defaultConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	svc := route53.NewFromConfig(defaultConfig)
	input := &route53.ListHostedZonesByNameInput{}

	input.DNSName = &hostedZone
	result, err := svc.ListHostedZonesByName(ctx, input)
	if err != nil {

		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err)

		return "", err
	}

	for _, v := range result.HostedZones {
		if hostedZone != *v.Name {
			return "", err
		} else {
			return *v.Id, nil
		}
	}
	return "", err
}

// checkIfAliasOrARecordExist

func checkIfAliasOrARecordExist(hostedZoneID string, LbToManage *awslb.AWSLb) bool {
	ctx := context.Background()
	defaultConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return false
	}
	svc := route53.NewFromConfig(defaultConfig)

	ghzi := &route53.GetHostedZoneInput{}
	ghzi.Id = &hostedZoneID
	r, err := svc.GetHostedZone(ctx, ghzi)
	if err != nil {
		fmt.Println(err)

		return false
	}

	rspg := route53.NewListResourceRecordSetsPaginator(svc, &route53.ListResourceRecordSetsInput{HostedZoneId: r.HostedZone.Id})

	for rspg.HasMorePages() {
		output, err := rspg.NextPage(context.TODO())
		if err != nil {
			fmt.Println(err)
			return false
		}
		for _, v := range output.ResourceRecordSets {
			if v.AliasTarget != nil {
				if *v.AliasTarget.DNSName == strings.ToLower(LbToManage.LBDNS+".") {

					recordWeManage := weManageThis(hostedZoneID, LbToManage)
					if recordWeManage != "" {
						if recordWeManage != LbToManage.TagValue {
							updateAlias(hostedZoneID, recordWeManage, LbToManage)
						} else {
							fmt.Println("No need to Update the records. LB and Route53 are in Sync.")
						}
					}
					return true
				}
			}
			// This means an A record exists for this hostedzone
			if *v.Name == LbToManage.TagValue+"." && v.Type == "A" {

				fmt.Println(LbToManage.TagValue + aRecordExist)
				return true
			}
		}
	}
	return false
}

// To create Alias in Route53

func createAlias(hostedZoneID string, LbToManage *awslb.AWSLb) error {
	smallHostedZoneID := *getSmallhostedZoneID(hostedZoneID)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}

	client := route53.NewFromConfig(cfg)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(LbToManage.TagValue),
						Type: types.RRTypeA,
						AliasTarget: &types.AliasTarget{
							DNSName:              aws.String(LbToManage.LBDNS),
							HostedZoneId:         aws.String(LbToManage.LBHostedZoneId),
							EvaluateTargetHealth: *aws.Bool(false),
						},
					},
				},
			},
		},
		HostedZoneId: aws.String(smallHostedZoneID),
	}

	_, err = client.ChangeResourceRecordSets(context.TODO(), params)
	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}
	fmt.Println("Created Alias: ", LbToManage.LBDNS)
	return nil
}

func createHostedZone(hostedZone string, LbToManage *awslb.AWSLb) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}
	t := time.Now()
	callerReference := t.Format("20060102150405")

	client := route53.NewFromConfig(cfg)
	params := &route53.CreateHostedZoneInput{
		CallerReference: &callerReference,
		Name:            &hostedZone,
		HostedZoneConfig: &types.HostedZoneConfig{
			Comment:     aws.String(hostedZoneDescription),
			PrivateZone: false,
		},
	}

	output, err := client.CreateHostedZone(context.TODO(), params)
	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}

	fmt.Println("Created Hosted Zone", *output.HostedZone.Name)
	createAlias(*output.HostedZone.Id, LbToManage)
	createHostedZoneTags(*output.HostedZone.Id, LbToManage)
	return nil
}

// This tags only records that we created where HostedZones Already exist
func createRecordOnlyTags(hostedZoneID string, LbToManage *awslb.AWSLb) {
	smallHostedZoneID := *getSmallhostedZoneID(hostedZoneID)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	client := route53.NewFromConfig(cfg)

	params := &route53.ChangeTagsForResourceInput{
		ResourceId:   &smallHostedZoneID,
		ResourceType: types.TagResourceTypeHostedzone,
		AddTags: []types.Tag{
			{Key: aws.String(aliasKey), Value: &LbToManage.LBDNS},
			{Key: aws.String(recordTag), Value: &LbToManage.TagValue},
		},
	}

	_, err = client.ChangeTagsForResource(context.TODO(), params)
	if err != nil {
		fmt.Println(err)

	}

}

// This one tags HostedZones that we created along with the records
func createHostedZoneTags(hostedZoneID string, LbToManage *awslb.AWSLb) {
	smallHostedZoneID := *getSmallhostedZoneID(hostedZoneID)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	client := route53.NewFromConfig(cfg)

	params := &route53.ChangeTagsForResourceInput{
		ResourceId:   &smallHostedZoneID,
		ResourceType: types.TagResourceTypeHostedzone,
		AddTags: []types.Tag{{Key: aws.String(dnsServiceManagedTagKey), Value: aws.String(dnsServiceManagedTagValue)},
			{Key: aws.String(aliasKey), Value: &LbToManage.LBDNS},
			{Key: aws.String(recordTag), Value: &LbToManage.TagValue},
		},
	}

	_, err = client.ChangeTagsForResource(context.TODO(), params)
	if err != nil {
		fmt.Println(err)

	}

}

func updateAlias(hostedZoneID, recordWeManage string, LbToManage *awslb.AWSLb) error {
	splitHostedZoneID := *getSmallhostedZoneID(hostedZoneID)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}

	client := route53.NewFromConfig(cfg)

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionDelete,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(recordWeManage),
						Type: types.RRTypeA,
						AliasTarget: &types.AliasTarget{
							DNSName:              aws.String(LbToManage.LBDNS),
							HostedZoneId:         aws.String(LbToManage.LBHostedZoneId),
							EvaluateTargetHealth: *aws.Bool(false),
						},
					},
				},
			},
		},
		HostedZoneId: aws.String(splitHostedZoneID),
	}

	_, err = client.ChangeResourceRecordSets(context.TODO(), params)

	if err != nil {
		// handle error
		fmt.Println(err)
		return err
	}
	fmt.Println("Deleted alias: ", recordWeManage)
	createAlias(hostedZoneID, LbToManage)
	createRecordOnlyTags(hostedZoneID, LbToManage)
	return nil
}

// Check if we manage this HostedZone
func weManageThis(hostedZoneID string, LbToManage *awslb.AWSLb) string {
	smallHostedZoneID := *getSmallhostedZoneID(hostedZoneID)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		// handle error
		fmt.Println(err)
		return ""
	}
	client := route53.NewFromConfig(cfg)
	params := &route53.ListTagsForResourceInput{
		ResourceId:   &smallHostedZoneID,
		ResourceType: types.TagResourceTypeHostedzone,
	}
	output, err := client.ListTagsForResource(context.TODO(), params)
	if err != nil {
		// handle error
		fmt.Println(err)
		return ""
	}
	existingRecord := ""
	for _, v := range output.ResourceTagSet.Tags {
		if *v.Key == recordTag {
			existingRecord = *v.Value

			fmt.Println("existingRecord", existingRecord)
		}
	}
	return existingRecord
}

func getSmallhostedZoneID(hostedZoneID string) *string {

	return &strings.Split(hostedZoneID, "/")[2]
}
