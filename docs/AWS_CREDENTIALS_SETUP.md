# AWS Credentials Setup for Snoozebot

This guide explains how to set up AWS credentials for use with Snoozebot's AWS cloud provider plugin.

## Prerequisites

- An active AWS account
- AWS CLI installed on your system
- Sufficient permissions to create IAM users and roles

## Step 1: Install AWS CLI

If you don't have AWS CLI installed, follow these instructions:

### macOS
```bash
brew install awscli
```

### Linux
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

### Windows
Download and install from: https://aws.amazon.com/cli/

## Step 2: Create an IAM User for Snoozebot

It's a best practice to create a dedicated IAM user for programmatic access:

1. Log in to the AWS Management Console
2. Navigate to the IAM service
3. Click "Users" in the left sidebar, then "Add user"
4. Set a username (e.g., "snoozebot-user")
5. Select "Programmatic access" as the access type

## Step 3: Attach Permissions to the IAM User

For testing purposes, you can attach the `AmazonEC2FullAccess` policy, but for production, consider creating a custom policy with minimal permissions:

1. On the permissions page, select "Attach existing policies directly"
2. Search for and select `AmazonEC2FullAccess`
3. Optionally, click "Create policy" to define a custom policy with minimal permissions
4. Review and create the user

After creating the user, you'll see the Access Key ID and Secret Access Key. Save these credentials securely; you won't be able to view the Secret Access Key again.

## Step 4: Set Up AWS CLI Profile

Run the following command and enter your credentials when prompted:

```bash
# Create a named profile for Snoozebot
aws configure --profile snoozebot
```

You'll be asked to provide:
- AWS Access Key ID
- AWS Secret Access Key
- Default region (e.g., us-west-2)
- Default output format (e.g., json)

This creates two files:
- `~/.aws/credentials` - Stores your access keys
- `~/.aws/config` - Stores your region and output preferences

## Step 5: Create a Custom Policy (Recommended for Production)

For better security, create a custom policy with only the permissions Snoozebot needs:

1. In the IAM console, go to "Policies" and click "Create policy"
2. Use the JSON editor and paste the following policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:StartInstances",
        "ec2:StopInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

3. Name the policy (e.g., "SnoozeBot-EC2-Policy") and create it
4. Attach this policy to your Snoozebot IAM user instead of the full access policy

## Step 6: Set Environment Variable for Tests

For running tests with AWS, set the following environment variable:

```bash
export AWS_PROFILE=snoozebot
```

You can add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it persistent:

```bash
echo 'export AWS_PROFILE=snoozebot' >> ~/.bashrc
source ~/.bashrc
```

## Step 7: Create AWS Configuration for Snoozebot

Create a configuration file for Snoozebot to use your AWS profile:

```bash
mkdir -p ~/.config/snoozebot
cat > ~/.config/snoozebot/aws-config.json << EOF
{
  "provider": "aws",
  "credentials": {
    "profileName": "snoozebot",
    "region": "us-west-2"
  }
}
EOF
```

## Verifying Your Setup

To verify your credentials are set up correctly, run:

```bash
# List EC2 instances using your profile
aws ec2 describe-instances --profile snoozebot
```

If this command returns without error (even if you don't have any instances), your credentials are working.

## Additional Security Considerations

1. **Use IAM Roles for EC2**: When running Snoozebot on EC2, use IAM roles instead of access keys.

2. **Regular Key Rotation**: Rotate your access keys regularly (every 90 days is a common practice).

3. **Enable MFA**: Enable multi-factor authentication for the IAM user.

4. **Use a VPC Endpoint**: Consider using a VPC endpoint for EC2 to keep traffic within the AWS network.

5. **Monitor with CloudTrail**: Enable AWS CloudTrail to audit API calls made with your credentials.

## Troubleshooting

### Error: "Unable to locate credentials"
- Ensure you've configured the AWS CLI correctly
- Check that the credentials file exists at `~/.aws/credentials`
- Verify you're using the correct profile name

### Error: "Access Denied"
- Check if the IAM user has the necessary permissions
- Verify that the Access Key ID and Secret Access Key are correct
- Ensure the IAM user is active and not deleted

### Region-specific errors
- Make sure you're specifying the correct region for your resources
- Some resources might be region-specific

## Using Temporary Credentials with AWS STS (Advanced)

For enhanced security, you can use AWS Security Token Service (STS) to obtain temporary credentials:

```bash
# Assume a role with temporary credentials
aws sts assume-role \
    --role-arn arn:aws:iam::123456789012:role/SnoozeBot-Role \
    --role-session-name snoozebot-session \
    --profile snoozebot
```

## Next Steps

Once your AWS credentials are set up, you can proceed with configuring the Snoozebot AWS cloud provider plugin in your application.

For more details on AWS IAM and security best practices, see the [AWS IAM documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/introduction.html).