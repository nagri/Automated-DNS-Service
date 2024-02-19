package awslb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// This will Scan and report AWS LBs from where we'll read the tags

type AWSLb struct {
	LBARN          string `json:"lb_arn"`
	LBDNS          string `json:"lb_dns"`
	TagKey         string `json:"tag_key"`
	TagValue       string `json:"tag_value"`
	LBHostedZoneId string `json:"lb_hosted_zone_id"`
}
type KnownAWSLBs struct {
	AWSLBs []AWSLb `json:"aws_lbs"`
}

// This is the tag we look for in the LB if it will be using this service
const automtedDNSTag = "automated-dns"

func ScanAWSLB() (*KnownAWSLBs, error) {

	var awsLBs KnownAWSLBs
	defaultConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	svc := elasticloadbalancingv2.NewFromConfig(defaultConfig)
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{},
	}
	ctx := context.Background()
	result, err := svc.DescribeLoadBalancers(ctx, input)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for _, v := range result.LoadBalancers {
		var LBArnList []string
		var awsLB AWSLb

		awsLB.LBHostedZoneId = *v.CanonicalHostedZoneId
		awsLB.LBDNS = *v.DNSName
		awsLB.LBARN = *v.LoadBalancerArn
		LBArnList = append(LBArnList, *v.LoadBalancerArn)
		tags := &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: LBArnList,
		}

		tagResult, err := svc.DescribeTags(ctx, tags)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		weManageThisLB := false
		for _, v := range tagResult.TagDescriptions {
			for _, tag := range v.Tags {
				if *tag.Key == automtedDNSTag {
					weManageThisLB = true
					awsLB.TagKey = *tag.Key
					awsLB.TagValue = *tag.Value
				}
			}
		}
		if weManageThisLB {
			awsLBs.AWSLBs = append(awsLBs.AWSLBs, awsLB)
		}
	}
	return &awsLBs, nil
}
