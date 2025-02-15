# DataBackup-day10-

#Introduction
This is a Sports Data Backup project for NBA Game Day. The project is designed to backup the data from the NBA Game Day API. The data is stored in a database and can be accessed by the user.

#System Design

#setup s3 bucket permisson

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
