# Setting Up Cloud Provider Test Environments for Snoozebot

This guide provides instructions for setting up test environments for AWS, Azure, and GCP to use with Snoozebot's integration and live tests.

## Overview

For effective testing of Snoozebot's cloud provider functionality, we need isolated and controlled test environments in each cloud platform. This document outlines the setup process for:

1. Credentials Configuration
2. Resource Creation
3. Test Instance Configuration
4. Security Best Practices

## Prerequisites

- Active accounts in AWS, Azure, and/or GCP
- Appropriate permissions to create and manage resources
- Cloud provider CLIs installed and configured:
  - AWS CLI
  - Azure CLI
  - Google Cloud SDK

## Automated Setup

We've provided scripts to help automate the setup process:

```bash
# Main script that guides you through all providers
./scripts/setup_cloud_credentials.sh

# Individual provider scripts
./scripts/setup_azure_credentials.sh
./scripts/setup_gcp_credentials.sh
# AWS typically uses existing AWS CLI configuration
```

## Manual Setup Instructions

### 1. AWS Test Environment

#### 1.1 Credentials Setup

If you already have AWS credentials set up through the AWS CLI, you can use an existing profile or create a new one:

```bash
# Create a new profile for testing
aws configure --profile snoozebot-test
```

#### 1.2 Creating a Test VPC and Subnet

```bash
# Create a VPC for testing
VPC_ID=$(aws ec2 create-vpc \
    --cidr-block 10.0.0.0/16 \
    --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=snoozebot-test-vpc}]' \
    --query Vpc.VpcId --output text)

# Create a subnet
SUBNET_ID=$(aws ec2 create-subnet \
    --vpc-id $VPC_ID \
    --cidr-block 10.0.1.0/24 \
    --availability-zone us-west-2a \
    --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=snoozebot-test-subnet}]' \
    --query Subnet.SubnetId --output text)

# Create an internet gateway
IGW_ID=$(aws ec2 create-internet-gateway \
    --tag-specifications 'ResourceType=internet-gateway,Tags=[{Key=Name,Value=snoozebot-test-igw}]' \
    --query InternetGateway.InternetGatewayId --output text)

# Attach the internet gateway to the VPC
aws ec2 attach-internet-gateway --vpc-id $VPC_ID --internet-gateway-id $IGW_ID

# Create a route table
ROUTE_TABLE_ID=$(aws ec2 create-route-table \
    --vpc-id $VPC_ID \
    --tag-specifications 'ResourceType=route-table,Tags=[{Key=Name,Value=snoozebot-test-rtb}]' \
    --query RouteTable.RouteTableId --output text)

# Create a route to the internet
aws ec2 create-route --route-table-id $ROUTE_TABLE_ID --destination-cidr-block 0.0.0.0/0 --gateway-id $IGW_ID

# Associate the route table with the subnet
aws ec2 associate-route-table --subnet-id $SUBNET_ID --route-table-id $ROUTE_TABLE_ID
```

#### 1.3 Creating a Security Group

```bash
# Create a security group
SG_ID=$(aws ec2 create-security-group \
    --group-name snoozebot-test-sg \
    --description "Security group for Snoozebot testing" \
    --vpc-id $VPC_ID \
    --query GroupId --output text)

# Add an SSH ingress rule
aws ec2 authorize-security-group-ingress \
    --group-id $SG_ID \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0
```

#### 1.4 Creating a Test Instance

```bash
# Create a key pair (if you don't already have one)
aws ec2 create-key-pair \
    --key-name snoozebot-test-key \
    --query "KeyMaterial" \
    --output text > snoozebot-test-key.pem
chmod 400 snoozebot-test-key.pem

# Launch a test instance
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id ami-0c55b159cbfafe1f0 \
    --instance-type t2.micro \
    --key-name snoozebot-test-key \
    --security-group-ids $SG_ID \
    --subnet-id $SUBNET_ID \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=snoozebot-test-instance}]' \
    --query Instances[0].InstanceId --output text)

# Wait for the instance to be running
aws ec2 wait instance-running --instance-ids $INSTANCE_ID

# Save the instance ID for later use
echo $INSTANCE_ID > snoozebot-test-instance-id.txt
```

### 2. Azure Test Environment

#### 2.1 Credentials Setup

Follow the instructions in [AZURE_CREDENTIALS_SETUP.md](AZURE_CREDENTIALS_SETUP.md) or use the automated script:

```bash
./scripts/setup_azure_credentials.sh
```

#### 2.2 Creating a Resource Group

```bash
# Set variables
RESOURCE_GROUP="snoozebot-test-rg"
LOCATION="eastus"

# Create a resource group
az group create --name $RESOURCE_GROUP --location $LOCATION
```

#### 2.3 Creating a Virtual Network

```bash
# Create a vnet and subnet
az network vnet create \
    --resource-group $RESOURCE_GROUP \
    --name snoozebot-test-vnet \
    --address-prefix 10.0.0.0/16 \
    --subnet-name snoozebot-test-subnet \
    --subnet-prefix 10.0.0.0/24
    
# Create a network security group
az network nsg create \
    --resource-group $RESOURCE_GROUP \
    --name snoozebot-test-nsg

# Create an SSH rule
az network nsg rule create \
    --resource-group $RESOURCE_GROUP \
    --nsg-name snoozebot-test-nsg \
    --name allow-ssh \
    --protocol tcp \
    --priority 1000 \
    --destination-port-range 22 \
    --access allow
```

#### 2.4 Creating a Test VM

```bash
# Create a public IP address
az network public-ip create \
    --resource-group $RESOURCE_GROUP \
    --name snoozebot-test-ip \
    --allocation-method Dynamic

# Create a network interface
az network nic create \
    --resource-group $RESOURCE_GROUP \
    --name snoozebot-test-nic \
    --vnet-name snoozebot-test-vnet \
    --subnet snoozebot-test-subnet \
    --network-security-group snoozebot-test-nsg \
    --public-ip-address snoozebot-test-ip

# Create a VM
az vm create \
    --resource-group $RESOURCE_GROUP \
    --name snoozebot-test-vm \
    --nics snoozebot-test-nic \
    --image UbuntuLTS \
    --admin-username azureuser \
    --generate-ssh-keys

# Save the VM name for later use
echo "snoozebot-test-vm" > snoozebot-test-vm-name.txt
```

### 3. GCP Test Environment

#### 3.1 Credentials Setup

Follow the instructions in [GCP_CREDENTIALS_SETUP.md](GCP_CREDENTIALS_SETUP.md) or use the automated script:

```bash
./scripts/setup_gcp_credentials.sh
```

#### 3.2 Creating a VPC Network

```bash
# Set variables
PROJECT_ID=$(gcloud config get-value project)
NETWORK_NAME="snoozebot-test-network"
REGION="us-central1"
ZONE="us-central1-a"

# Create a custom VPC network
gcloud compute networks create $NETWORK_NAME --subnet-mode=custom

# Create a subnet
gcloud compute networks subnets create snoozebot-test-subnet \
    --network=$NETWORK_NAME \
    --region=$REGION \
    --range=10.0.0.0/24

# Create a firewall rule to allow SSH
gcloud compute firewall-rules create snoozebot-test-allow-ssh \
    --direction=INGRESS \
    --priority=1000 \
    --network=$NETWORK_NAME \
    --action=ALLOW \
    --rules=tcp:22 \
    --source-ranges=0.0.0.0/0
```

#### 3.3 Creating a Test Instance

```bash
# Create a test VM instance
gcloud compute instances create snoozebot-test-instance \
    --zone=$ZONE \
    --machine-type=e2-micro \
    --subnet=snoozebot-test-subnet \
    --network-tier=PREMIUM \
    --image-family=debian-10 \
    --image-project=debian-cloud \
    --boot-disk-size=10GB \
    --boot-disk-type=pd-standard

# Save the instance name for later use
echo "snoozebot-test-instance" > snoozebot-test-instance-name.txt
```

## Configuration for Snoozebot Tests

### Setting Environment Variables

Set the following environment variables before running tests:

```bash
# AWS
export AWS_PROFILE=snoozebot-test
export SNOOZEBOT_TEST_INSTANCE_ID=$(cat snoozebot-test-instance-id.txt)

# Azure
export AZURE_PROFILE=snoozebot
export SNOOZEBOT_AZURE_VM_NAME=$(cat snoozebot-test-vm-name.txt)
export SNOOZEBOT_AZURE_RESOURCE_GROUP=snoozebot-test-rg

# GCP
export GCP_PROFILE=snoozebot
export GOOGLE_APPLICATION_CREDENTIALS=~/.config/gcloud/snoozebot/snoozebot-gcp-key.json
export SNOOZEBOT_GCP_INSTANCE_NAME=$(cat snoozebot-test-instance-name.txt)
export SNOOZEBOT_GCP_ZONE=us-central1-a

# Enable live tests
export SNOOZEBOT_LIVE_TESTS=true

# Enable start/stop tests (use with caution)
export SNOOZEBOT_TEST_START_STOP=true
```

## Running the Tests

With your environment set up and variables configured, you can run the tests:

```bash
# Run all tests
go test ./test/... -v

# Run only AWS tests
go test ./test/aws/... -v

# Run only Azure tests
go test ./test/azure/... -v

# Run only GCP tests
go test ./test/gcp/... -v
```

## Cleanup

After testing, clean up your resources to avoid unnecessary charges:

### AWS Cleanup

```bash
# Terminate the test instance
aws ec2 terminate-instances --instance-ids $INSTANCE_ID

# Wait for the instance to be terminated
aws ec2 wait instance-terminated --instance-ids $INSTANCE_ID

# Delete the security group
aws ec2 delete-security-group --group-id $SG_ID

# Detach the internet gateway
aws ec2 detach-internet-gateway --internet-gateway-id $IGW_ID --vpc-id $VPC_ID

# Delete the internet gateway
aws ec2 delete-internet-gateway --internet-gateway-id $IGW_ID

# Delete the route table
aws ec2 delete-route-table --route-table-id $ROUTE_TABLE_ID

# Delete the subnet
aws ec2 delete-subnet --subnet-id $SUBNET_ID

# Delete the VPC
aws ec2 delete-vpc --vpc-id $VPC_ID
```

### Azure Cleanup

```bash
# Delete the entire resource group
az group delete --name $RESOURCE_GROUP --yes
```

### GCP Cleanup

```bash
# Delete the test instance
gcloud compute instances delete snoozebot-test-instance --zone=$ZONE --quiet

# Delete the firewall rule
gcloud compute firewall-rules delete snoozebot-test-allow-ssh --quiet

# Delete the subnet
gcloud compute networks subnets delete snoozebot-test-subnet --region=$REGION --quiet

# Delete the network
gcloud compute networks delete $NETWORK_NAME --quiet
```

## Security Best Practices

1. **Use Minimal Permissions**: The service principals and IAM roles should have only the permissions needed for testing.

2. **Isolate Test Resources**: Use dedicated VPCs/Networks and resource groups for testing.

3. **Clean Up After Testing**: Always delete test resources after you're done to avoid ongoing charges.

4. **Protect Credentials**: Never commit credentials or key files to source control.

5. **Use Firewalls**: Restrict access to test instances to your IP address when possible.

## Troubleshooting

### AWS Issues

- **Error: "VPC limit exceeded"**: Delete unused VPCs or request a limit increase.
- **SSH Connection Issues**: Check security group rules and key pairs.

### Azure Issues

- **Authentication Errors**: Verify service principal credentials and permissions.
- **Resource Creation Failures**: Check quota limits in your subscription.

### GCP Issues

- **API Not Enabled**: Ensure the Compute Engine API is enabled for your project.
- **Permission Denied**: Check service account roles and permissions.

## Next Steps

After setting up your test environments, proceed with implementing and running the cloud provider tests for Snoozebot.

For more information, refer to the specific documentation for each cloud provider:

- [AWS Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EC2_GetStarted.html)
- [Azure Documentation](https://docs.microsoft.com/en-us/azure/virtual-machines/)
- [GCP Documentation](https://cloud.google.com/compute/docs)