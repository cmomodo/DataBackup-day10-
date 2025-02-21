# 1. Delete EventBridge Rule
aws events remove-targets --rule sports-backup-rule --ids "1"
aws events delete-rule --name sports-backup-rule

# 2. Delete ECS Resources
# Stop any running tasks
aws ecs list-tasks --cluster sports-backup-cluster | \
aws ecs stop-task --cluster sports-backup-cluster --task $(jq -r '.taskArns[]')

# Delete the ECS cluster
aws ecs delete-cluster --cluster sports-backup-cluster

# Deregister task definition
aws ecs deregister-task-definition --task-definition sports-backup-task:1

# 3. Delete IAM Roles and Policies
# Detach policies from roles
aws iam detach-role-policy --role-name ecsEventsRole --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
aws iam delete-role-policy --role-name sports-backup-events-role --policy-name sports-backup-events-policy

# Delete roles
aws iam delete-role --role-name ecsEventsRole
aws iam delete-role --role-name sports-backup-task-execution-role

# 4. Delete CloudWatch Log Group
aws logs delete-log-group --log-group-name /ecs/sports-backup

# 5. Delete ECR Repository
aws ecr delete-repository --repository-name sports-backup --force

# 6. Delete VPC Resources
# Delete security group
aws ec2 delete-security-group --group-id sg-0e546ca47bfbc276a

# Delete subnet
aws ec2 delete-subnet --subnet-id subnet-08761086485cf7a83

# Delete VPC (replace with your VPC ID)
aws ec2 delete-vpc --vpc-id vpc-04446d2d411c2f955

# 7. Delete S3 Bucket
# First empty the bucket
aws s3 rm s3://json-store-acm --recursive
aws s3api delete-bucket --bucket json-store-acm

# 8. Delete DynamoDB Table
aws dynamodb delete-table --table-name SportsHighlights