# [Case Study] Deploying and Scaling an AWS ECS Cluster Using AWS CDK

## **Emulator Cluster Overview**
### **1. Web Server Emulator (EC2 Launch Type)**
- Simulates a web server that scales out when demand peaks.
- Every request to `/` starts a goroutine consuming **80% of vCPU** and some memory.
- Uses an **Application Load Balancer** and an **Auto Scaling Group** with a target scaling policy of **50% CPU utilization**.
- Minimum 1, Maximum 3 instances.

### **2. Standalone Task Emulator (Fargate Launch Type)**
- Similar to the web server emulator but runs as an on-demand task that terminates after execution.
- Logs can be inspected for performance analysis.

## **Instance & Task Scaling Policies**
- **Web Server Emulator:** Minimum 1, Maximum 3, **Target CPU Utilization: 50%**.
- **Task Emulator:** Runs as needed without a persistent scaling policy.

---

## **Project Setup**

### 1. Prerequisites:

Please install the following tools, and make sure you have a AWS Account

* Docker
* AWS CLI
* AWS CDK CLI

### **2. Configure Environment**
Open the `.env` file and fill in the required values, including your **AWS Account ID** and **AWS Region**.

---

### **3. Build and Push Docker Images to Amazon ECR**

#### **Login to AWS ECR**
```sh
aws ecr get-login-password | docker login --username AWS --password-stdin <aws-account-id>.dkr.ecr.<aws-region>.amazonaws.com
```

#### **Build and Push Task Emulator**
```sh
cd docker/task-emulator

docker build -t task-emulator -f dockerfile ../../

aws ecr create-repository --repository-name task-emulator

docker tag task-emulator <aws-account-id>.dkr.ecr.<aws-region>.amazonaws.com/task-emulator

docker push <aws-account-id>.dkr.ecr.<aws-region>.amazonaws.com/task-emulator
```

#### **Build and Push Web Server Emulator**
```sh
cd docker/web-server-emulator

docker build -t web-server-emulator -f dockerfile ../../

aws ecr create-repository --repository-name web-server-emulator

docker tag web-server-emulator <aws-account-id>.dkr.ecr.<aws-region>.amazonaws.com/web-server-emulator

docker push <aws-account-id>.dkr.ecr.<aws-region>.amazonaws.com/web-server-emulator
```

---

### **4. Deploy Using AWS CDK**
```sh
go mod tidy

cdk bootstrap

cdk deploy ScalingExperimentVpc

cdk deploy ScalingExperimentEmulatorsCluster
```

---

### **5. Verify Deployment**
1. Log in to the **AWS Management Console** and check if the following resources were created:
   - **Stack, ECS Cluster, VPC, Task Definitions, Services, and Load Balancer**.
2. Verify that **two EC2 instances** were created:
   - The first instance runs the **web emulator service**.
   - The second instance is provisioned because the first instance exceeds **50% CPU utilization**, triggering auto-scaling.
3. Check the ECS Cluster:
   - Ensure the **web emulator service** is running at least one task.

#### **Test the Web Emulator Endpoint**
1. Go to the **Load Balancer** in the AWS Console.
2. Copy the **DNS name** and paste it into your browser.
3. You should see an output similar to this:
   ```
   client ip: xxx.80.215.xxx
   Emulating running for 15 seconds: 80% CPU - 1 MB of memory.
   ```

---

### **6. Test Auto Scaling**
To simulate peak demand, run the following script in your browser's console:
```js
var interval = setInterval(() => fetch("<your lb dns here>"), 2000);
```
- Wait **approximately 2 minutes**, then check the ECS Cluster.
- The service should scale from **1 to 2 tasks**, running on two instances to handle the increased load.

#### **Stop the Load Simulation**
After inspecting the scaling behavior, stop the simulation by running:
```js
clearInterval(interval);
```
### 7. Cleanup

Clean AWS Resources:

```
cdk destroy ScalingExperimentEmulatorsCluster
cdk destroy ScalingExperimentVpc
```

The process should finish after 30 minutes to 1 hour, otherwise, there might be difficuties happened. In this case, you must try to delete the resources manually with the AWS Management Console.
* Go to AWS CloudFormation > Stacks. Here, you can clearly see which resources deletion process is idling with DELETE_IN_PROGRESS status. 
* Delete the idling resources accordingly. Then wait for AWS to do the rest.