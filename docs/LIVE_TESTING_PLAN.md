# Live Testing Plan for Cloud Providers

This document outlines the approach for testing Snoozebot against live cloud provider environments (AWS, Azure, GCP). It covers required resources, test configuration, and methods to minimize costs during testing.

## Testing Goals

1. Verify that Snoozebot can interact with real cloud environments
2. Validate cloud provider plugins against actual infrastructure
3. Ensure resource operations (list, start, stop) work correctly
4. Test error handling with real-world conditions
5. Measure performance under realistic scenarios

## Cloud Provider Setup

### AWS Test Environment

#### Required Resources
- 1-2 EC2 t2.micro instances (free tier eligible)
- Test VPC with single subnet
- IAM user with limited EC2 permissions
- Security group allowing SSH access from test location

#### Setup Steps
```bash
# Create test VPC and subnet
aws ec2 create-vpc --cidr-block 10.0.0.0/16 --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=snoozebot-test}]'
aws ec2 create-subnet --vpc-id <vpc-id> --cidr-block 10.0.1.0/24

# Create security group
aws ec2 create-security-group --group-name snoozebot-test --description "Snoozebot test security group" --vpc-id <vpc-id>
aws ec2 authorize-security-group-ingress --group-id <sg-id> --protocol tcp --port 22 --cidr 0.0.0.0/0

# Launch test instances
aws ec2 run-instances --image-id ami-0c55b159cbfafe1f0 --count 2 --instance-type t2.micro --key-name <key-name> --security-group-ids <sg-id> --subnet-id <subnet-id> --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=snoozebot-test}]'
```

#### Testing Credentials
Create a limited IAM user with permissions only to:
- DescribeInstances
- StartInstances
- StopInstances

### Azure Test Environment

#### Required Resources
- 1-2 B1s VMs (low cost)
- Resource group for testing
- Service principal with limited rights
- Network security group

#### Setup Steps
```bash
# Create resource group
az group create --name snoozebot-test --location eastus

# Create VM
az vm create \
  --resource-group snoozebot-test \
  --name snoozebot-test-vm \
  --image UbuntuLTS \
  --size B1s \
  --admin-username azureuser \
  --generate-ssh-keys \
  --tags "purpose=testing"

# Create service principal with limited rights
az ad sp create-for-rbac --name snoozebot-test-sp --role contributor --scopes /subscriptions/<subscription-id>/resourceGroups/snoozebot-test
```

### GCP Test Environment

#### Required Resources
- 1-2 e2-micro instances (free tier eligible)
- Dedicated test project
- Service account with minimal permissions
- VPC network with firewall rules

#### Setup Steps
```bash
# Create test project (if needed)
gcloud projects create snoozebot-test --name="Snoozebot Testing"

# Set project
gcloud config set project snoozebot-test

# Create VM instances
gcloud compute instances create snoozebot-test-1 snoozebot-test-2 \
  --machine-type=e2-micro \
  --zone=us-central1-a \
  --image-family=debian-11 \
  --image-project=debian-cloud \
  --tags=test

# Create service account
gcloud iam service-accounts create snoozebot-test-sa \
  --display-name="Snoozebot Test Service Account"

# Assign permissions
gcloud projects add-iam-policy-binding snoozebot-test \
  --member="serviceAccount:snoozebot-test-sa@snoozebot-test.iam.gserviceaccount.com" \
  --role="roles/compute.viewer"

gcloud projects add-iam-policy-binding snoozebot-test \
  --member="serviceAccount:snoozebot-test-sa@snoozebot-test.iam.gserviceaccount.com" \
  --role="roles/compute.instanceAdmin.v1"
```

## Test Scenarios

### Basic Functionality Tests

1. **Instance Discovery**
   - Test listing all instances
   - Verify instance metadata (ID, name, type, state)
   - Verify zone/region information

2. **Instance State Management**
   - Stop a running instance
   - Start a stopped instance
   - Verify state changes are properly detected

3. **Error Handling**
   - Test with invalid instance IDs
   - Test with revoked/expired credentials
   - Test network disruption scenarios

### Performance Tests

1. **Parallel Operations**
   - List instances from multiple providers simultaneously
   - Perform operations on multiple instances concurrently

2. **Response Time**
   - Measure time to list instances
   - Measure time to change instance state
   - Benchmark against direct API calls

### Security Tests

1. **Authentication Testing**
   - Verify API key authentication works with real credentials
   - Test behavior with invalid keys
   - Test expired credentials

2. **TLS Testing**
   - Verify secure communication with cloud APIs
   - Test certificate validation

## Test Execution

### Automated Testing Script

Create a script that:
1. Sets up cloud provider credentials from environment variables
2. Connects to each provider
3. Lists all instances
4. Performs operations on test instances
5. Verifies results and logs findings

```go
// Example test script structure
func TestAllProviders(t *testing.T) {
    // Only run in integration test mode
    if os.Getenv("SNOOZEBOT_LIVE_TESTS") != "true" {
        t.Skip("Skipping live provider tests")
    }
    
    // Run tests for each provider
    t.Run("AWS", testAWSProvider)
    t.Run("Azure", testAzureProvider)
    t.Run("GCP", testGCPProvider)
}

func testAWSProvider(t *testing.T) {
    // Setup AWS provider
    // Test listing instances
    // Test starting/stopping instances
    // Test error conditions
}
```

### Manual Test Plan

For manual testing, follow these steps:
1. Configure credentials for each provider
2. Start Snoozebot agent with plugin discovery enabled
3. Connect to the agent and list instances
4. Perform state change operations on test instances
5. Verify changes through cloud provider console

## Cost Management

### Controlling Test Costs

1. **Instance Schedule**
   - Start instances only during testing
   - Use automation to shut down instances after tests
   - Schedule tests during business hours only

2. **Resource Limits**
   - Use smallest viable instance types
   - Limit number of test instances to minimum needed
   - Delete unused resources immediately

3. **Billing Alerts**
   - Set up billing alerts/budgets in each cloud
   - Monitor costs daily during testing phase
   - Set hard limits where possible

### Cleanup Procedure

After each test session:

```bash
# AWS cleanup
aws ec2 terminate-instances --instance-ids <instance-id-1> <instance-id-2>

# Azure cleanup
az group delete --name snoozebot-test --yes

# GCP cleanup
gcloud compute instances delete snoozebot-test-1 snoozebot-test-2 --zone=us-central1-a --quiet
```

Create a cleanup script that runs after tests or on CI completion.

## Test Reporting

Generate a report after live testing with:
1. Test coverage metrics
2. Pass/fail status for each scenario
3. Performance measurements
4. Error logs and debugging information

## Prerequisites for Testing

- Active accounts on AWS, Azure, and GCP
- Payment method registered with each provider
- Permissions to create resources
- Ability to create service accounts/IAM users
- Network access to cloud APIs

## Resuming Testing

If testing is interrupted, follow these steps to resume:
1. Check cloud consoles for any running resources
2. Verify environment credentials are still valid
3. Review last completed test scenario
4. Resume from the next uncompleted scenario