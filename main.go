package main

import (
	awslb "Automated-DNS-Service/awsLB"
	r53 "Automated-DNS-Service/awsRoute53"
	utils "Automated-DNS-Service/utils"
	"fmt"
	"time"
)

func main() {

	for {

		LbsToManage, err := awslb.ScanAWSLB()
		if err != nil {
			fmt.Println(err)
		}

		for _, lb := range LbsToManage.AWSLBs {

			hostedZone := utils.GetHostedZoneString(lb.TagValue)

			fmt.Println("hostedZone: ", hostedZone)

			err := r53.ManageDNS(hostedZone, &lb)
			if err != nil {
				fmt.Println(err)
			}
		}

		time.Sleep(time.Second * 300)
	}
}
