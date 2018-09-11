package main

import (
	"fmt"
	"regexp"
	"strings"

	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Need search critera")
		os.Exit(1)
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		exitErrorf("failed to load config, %v", err)
	}

	searchString := os.Args[1]

	awsRegion := os.Getenv("AWS_REGION")

	cfg.Region = awsRegion
	svc := ec2.New(cfg)
	filter := []ec2.Filter{}

	rInstance := regexp.MustCompile(`^i-.*`)
	rIp := regexp.MustCompile(`^[0-9].*`)

	switch {
	case rInstance.MatchString(searchString): //instance id filter
		filter = []ec2.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []string{
					strings.Join([]string{searchString, "*"}, ""),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []string{
					string("running"),
				},
			},
		}

	case rIp.MatchString(searchString): // IP filter
		filter = []ec2.Filter{
			{
				Name: aws.String("private-ip-address"),
				Values: []string{
					strings.Join([]string{searchString, "*"}, ""),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []string{
					string("running"),
				},
			},
		}
	default: //Name filter
		filter = []ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []string{
					strings.Join([]string{"*", searchString, "*"}, ""),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []string{
					string("running"),
				},
			},
		}

	}

	params := &ec2.DescribeInstancesInput{
		Filters: filter,
	}
	req := svc.DescribeInstancesRequest(params)
	res, err := req.Send()
	if err != nil {
		exitErrorf("failed to describe instances, %s, %v", awsRegion, err)
	}

	for _, r := range res.Reservations {
		for _, i := range r.Instances {
			for _, t := range i.Tags {
				if *t.Key == "Name" {
					fmt.Println(*t.Value)
				}
			}
		}
	}
}
