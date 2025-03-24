package main

import (
	vpc "scaling_experiment/stack/vpc"
	emulator "scaling_experiment/stack/emulator-cluster"
	
	"os"
	"github.com/joho/godotenv"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	godotenv.Load()

	app := awscdk.NewApp(nil)

	vpcStack := vpc.NewVpcStack(app, "ScalingExperimentVpc", &vpc.VpcStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	emulator.NewEmulatorClusterStack(app, "ScalingExperimentEmulatorsCluster", &emulator.EmulatorStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Vpc: vpcStack.Vpc,
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region: jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
