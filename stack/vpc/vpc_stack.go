package cluster

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type VpcStackProps struct {
	StackProps awscdk.StackProps
}

type VpcStack struct {
	Stack	awscdk.Stack
	Vpc		awsec2.Vpc
}

func NewVpcStack(scope constructs.Construct, id string, props *VpcStackProps) *VpcStack {
	stack := &VpcStack{
		Stack: awscdk.NewStack(scope, &id, &props.StackProps),
	}

	stack.declareVpc()

	return stack
}

func (scope *VpcStack) declareVpc() {
	vpc := awsec2.NewVpc(scope.Stack, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
		ReservedAzs: jsii.Number(1),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("PrivateEgress"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	scope.Vpc = vpc
}

