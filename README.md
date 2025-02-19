# DataBackup-day10-

#Introduction
This is a Sports Data Backup project for NBA Game Day. The project is designed to backup the data from the NBA Game Day API. The data is stored in a database and can be accessed by the user.

# System Design

![System Design](/image/sports_lake_backup.png)

# Setup S3 Bucket Permission

```bash
aws s3api put-bucket-policy --bucket nba-game-day-data-backup --policy file://s3_policy.json
```

#Create MediaConvert endpoint

```bash
aws mediaconvert describe-endpoints
```

#run bash script

```bash
chmod +x vcp_setup.sh
./vcp_setup.sh
```

step 3: load environment variables

```bash

set -a
source .env
set +a
```

optional: confirm the environment variables are loaded

```bash
echo $AWS_ACCOUNT_ID
/ecs/sports-backup
sports-backup-task
449095351082
```

step 4: Generate final json files from template files

1. ECS Task Definition JSON File

```bash
envsubst < taskdef.template.json > taskdef.json
```

2. s3 dynamodbpolicy

```bash
envsubst < s3_dynamodb_policy.template.json > s3_dynamodb_policy.json
```

3. ECS Target

```bash
envsubst < ecsTarget.template.json > ecsTarget.json
```

4. ECS Events Role Policy

```bash
envsubst < ecseventsrole-policy.template.json > ecseventsrole-policy.json
```

Step 5: Build and Push Docker Image

##### create ecr repository

```bash
aws ecr create-repository \
    --repository-name sports-lake-ecr \
    --region us-east-1
```

##### login to docker registry

```bash
$(aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 765095351082.dkr.ecr.us-east-1.amazonaws.com)
```

#### build image

```bash
docker build -t sports-lake .
```

##### tag image with ECR URI

```bash
docker tag sports-lake:latest 765095351082.dkr.ecr.us-east-1.amazonaws.com/sports-lake-ecr:latest
```

##### push image to ECR

```bash
docker push 765095351082.dkr.ecr.us-east-1.amazonaws.com/sports-lake-ecr:latest
```

Step 6: Creation Of resources:
aws loggroup creation:

```aws logs create-log-group --log-group-name /ecs/sports-backup --region us-east-1

```

register task definition:

```bash
aws ecs register-task-definition --cli-input-json file://taskdef.json
```

attach s3 and dynamodb policy to role:

```bash
aws iam attach-role-policy --role-name sports-backup-task-execution-role --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

create events role:

```bash
aws iam create-role --role-name sports-backup-events-role --assume-role-policy-document file://ecseventsrole-trust.json
```

attach events role policy:

```bash
aws iam put-role-policy --role-name sports-backup-events-role --policy-name sports-backup-events-policy --policy-document file://ecseventsrole-policy.json
```

STEP 7: Create EventBridge Rule

```bash

aws events put-rule --name sports-backup-rule --schedule-expression 'rate(1 hour)' --region us-east-1
```

2. Add target to rule

```bash
aws events put-targets --rule sports-backup-rule --targets file://ecsTarget.json
```
