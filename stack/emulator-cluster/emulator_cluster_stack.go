package emulator

import (
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	scaling "github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	ec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	ecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	ecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	elb "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	constructs "github.com/aws/constructs-go/constructs/v10"
	jsii "github.com/aws/jsii-runtime-go"
)

const (
	WebProvider		= "web_server_emulator_capacity_provider"
	TaskProvider	= "task_emulator_capacity_provider"

	WebSG	= "web_server_security_group"
	TaskSG	= "standalone_task_security_group"
)

type EmulatorStackProps struct {
	StackProps	cdk.StackProps
	Vpc			ec2.Vpc
}

type EmulatorStack struct {
	Stack	cdk.Stack
	Props	*EmulatorStackProps

	Cluster				ecs.Cluster
	CapacityProviders	map[string]ecs.AsgCapacityProvider
	SecurityGroups		map[string]ec2.SecurityGroup
}

func NewEmulatorClusterStack(scope constructs.Construct, id string, props *EmulatorStackProps) *EmulatorStack {
	stack := &EmulatorStack{
		Stack: cdk.NewStack(scope, &id, &props.StackProps),
		Props: props,
		CapacityProviders: make(map[string]ecs.AsgCapacityProvider),
		SecurityGroups: make(map[string]ec2.SecurityGroup),
	}

	stack.declareSecurityGroups()
	stack.declareCluster()
	stack.declareWebServerEmulator()

	return stack
}

func (scope *EmulatorStack) declareSecurityGroups() {
	// Web server emulator security group
	webSG := ec2.NewSecurityGroup(scope.Stack, jsii.String("WebSecurityGroup"), &ec2.SecurityGroupProps{
		Vpc: scope.Props.Vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description: jsii.String("Security group for web server allowing HTTP, HTTPS, and SSH."),
	})
	webSG.AddIngressRule(ec2.Peer_AnyIpv4(), ec2.Port_Tcp(jsii.Number(80)), jsii.String("Allow HTTP"), nil)
	webSG.AddIngressRule(ec2.Peer_AnyIpv4(), ec2.Port_Tcp(jsii.Number(443)), jsii.String("Allow HTTPS"), nil)
	webSG.AddIngressRule(ec2.Peer_AnyIpv4(), ec2.Port_Tcp(jsii.Number(22)), jsii.String("Allow SSH"), nil)
	scope.SecurityGroups[WebSG] = webSG
	// Standalone task emulator
	taskSG := ec2.NewSecurityGroup(scope.Stack, jsii.String("TaskSecurityGroup"), &ec2.SecurityGroupProps{
		Vpc: scope.Props.Vpc,
		AllowAllOutbound: jsii.Bool(true),
		Description: jsii.String("Security group for standalone task. No inbound."),
	})
	scope.SecurityGroups[TaskSG] = taskSG
}

func (scope *EmulatorStack) declareWebServerEmulator() {
	// task definition
	taskDefinition := ecs.NewTaskDefinition(scope.Stack, jsii.String("WebTaskDefinition"), &ecs.TaskDefinitionProps{
		Cpu: jsii.String("0.75 vCPU"),
		MemoryMiB: jsii.String("0.75 GB"),
		Compatibility: ecs.Compatibility_EC2,
		NetworkMode: ecs.NetworkMode_AWS_VPC,
	})
	// container
	repo := ecr.Repository_FromRepositoryName(scope.Stack, jsii.String("WebRepo"), jsii.String("web-server-emulator"))
	taskDefinition.AddContainer(jsii.String("WebContainer"), &ecs.ContainerDefinitionOptions{
		Image: ecs.ContainerImage_FromEcrRepository(repo, nil),
		Cpu: jsii.Number(768),
		MemoryLimitMiB: jsii.Number(768), // hard limit
		MemoryReservationMiB: jsii.Number(512), // soft limit
		Essential: jsii.Bool(true),
		PortMappings: &[]*ecs.PortMapping{
			{ Name: jsii.String("http"), HostPort: jsii.Number(80), ContainerPort: jsii.Number(80), AppProtocol: ecs.AppProtocol_Http() },
		},
		EnableRestartPolicy: jsii.Bool(true),
		Command: &[]*string{
			jsii.String("/web-server-emulator"),
			jsii.String("--cpu"),
			jsii.String("80"),
			jsii.String("--mem"),
			jsii.String("1"),
			jsii.String("--runtime"),
			jsii.String("15"),
		},
		Logging: ecs.AwsLogDriver_AwsLogs(&ecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("ecs"),
			Mode: ecs.AwsLogDriverMode_NON_BLOCKING,
			MaxBufferSize: cdk.Size_Bytes(jsii.Number(25 * 1024 * 1024)), // 25mb
		}),
	})
	// web service
	service := ecs.NewEc2Service(scope.Stack, jsii.String("WebService"), &ecs.Ec2ServiceProps{
		Cluster: scope.Cluster,
		TaskDefinition: taskDefinition,
		DesiredCount: jsii.Number(1),
		AvailabilityZoneRebalancing: ecs.AvailabilityZoneRebalancing_ENABLED,
		SecurityGroups: &[]ec2.ISecurityGroup{
			scope.SecurityGroups[WebSG],
		},
		EnableExecuteCommand: jsii.Bool(true),
		CircuitBreaker: &ecs.DeploymentCircuitBreaker{
			Enable: jsii.Bool(true),
			Rollback: jsii.Bool(true),
		},
	})
	// load balancer for web service
	lb := elb.NewApplicationLoadBalancer(scope.Stack, jsii.String("WebLoadBalancer"), &elb.ApplicationLoadBalancerProps{
		Vpc: scope.Props.Vpc,
		InternetFacing: jsii.Bool(true),
	})
	listener := lb.AddListener(jsii.String("WebServiceListener"), &elb.BaseApplicationListenerProps{
		Port: jsii.Number(80),
	})
	listener.AddTargets(jsii.String("WebServiceTargetGroup"), &elb.AddApplicationTargetsProps{
		Port: jsii.Number(80),
		Targets: &[]elb.IApplicationLoadBalancerTarget{
			service.LoadBalancerTarget(&ecs.LoadBalancerTargetOptions{
				ContainerName: jsii.String(*taskDefinition.DefaultContainer().ContainerName()),
				ContainerPort: jsii.Number(80),
			}),
		},
		HealthCheck: &elb.HealthCheck{
			Enabled: jsii.Bool(true),
			// Interval: cdk.Duration_Seconds(jsii.Number(15)),
			// Timeout: cdk.Duration_Seconds(jsii.Number(10)),
			HealthyThresholdCount: jsii.Number(5),
			UnhealthyThresholdCount: jsii.Number(3),
		},
	})
	// target auto scaling (cpu ultilization)
	serviceScaling := service.AutoScaleTaskCount(&awsapplicationautoscaling.EnableScalingProps{
		MinCapacity: jsii.Number(1),
		MaxCapacity: jsii.Number(3),
	})
	serviceScaling.ScaleOnCpuUtilization(jsii.String("WebServiceCpuUtilScalingPolicy"), &ecs.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(50),
	})
}

func (scope *EmulatorStack) declareCluster() {
	cluster := ecs.NewCluster(scope.Stack, jsii.String("Cluster"), &ecs.ClusterProps{
		Vpc: scope.Props.Vpc,
		EnableFargateCapacityProviders: jsii.Bool(true),
	})
	// Web server emulator capacity provider
	webAsg := scaling.NewAutoScalingGroup(scope.Stack, jsii.String("WebAsg"), &scaling.AutoScalingGroupProps{
		Vpc: scope.Props.Vpc,
		InstanceType: ec2.NewInstanceType(jsii.String("t2.micro")), 
		MachineImage: ecs.EcsOptimizedImage_AmazonLinux2(ecs.AmiHardwareType_STANDARD, nil),
		MinCapacity: jsii.Number(1),
		MaxCapacity: jsii.Number(3),
		KeyPair: ec2.KeyPair_FromKeyPairName(scope.Stack, jsii.String("WebAsgKeyPair"), jsii.String("aws-huunv-general-purpose")),
		InstanceMonitoring: scaling.Monitoring_DETAILED,
	})
	webProvider := ecs.NewAsgCapacityProvider(scope.Stack, jsii.String("WebCapacityProvider"), &ecs.AsgCapacityProviderProps{
		AutoScalingGroup: webAsg,
		InstanceWarmupPeriod: jsii.Number(300),
		TargetCapacityPercent: jsii.Number(50),
	})
	cluster.AddAsgCapacityProvider(webProvider, nil)
	// Standalone task emulator capacity provider
	// taskAsg := scaling.NewAutoScalingGroup(scope.Stack, jsii.String("TaskEmulator"), &scaling.AutoScalingGroupProps{
	// 	Vpc: scope.Props.Vpc,
	// 	InstanceType: ec2.NewInstanceType(jsii.String("t2.small")), 
	// 	MachineImage: ecs.EcsOptimizedImage_AmazonLinux2(ecs.AmiHardwareType_STANDARD, nil),
	// 	MinCapacity: jsii.Number(1),
	// 	MaxCapacity: jsii.Number(2),
	// 	KeyPair: ec2.KeyPair_FromKeyPairName(scope.Stack, jsii.String("AsgKeyPair"), jsii.String("aws-huunv-general-purpose")),
	// 	InstanceMonitoring: scaling.Monitoring_DETAILED,
	// })
	// taskProvider := ecs.NewAsgCapacityProvider(scope.Stack, jsii.String("TaskEmulatorCapacityProvider"), &ecs.AsgCapacityProviderProps{
	// 	AutoScalingGroup: taskAsg,
	// 	InstanceWarmupPeriod: jsii.Number(300),
	// })
	// cluster.AddAsgCapacityProvider(taskProvider, nil)
	
	scope.CapacityProviders[WebProvider] = webProvider
	// scope.CapacityProviders[TaskEmulatorProvider] = taskProvider
	scope.Cluster = cluster
}
