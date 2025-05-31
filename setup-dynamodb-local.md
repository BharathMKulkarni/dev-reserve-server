# Setting up DynamoDB Local

## Option 1: Using Docker (Recommended)

If you have Docker installed, this is the simplest approach:

```powershell
# Pull the DynamoDB Local image
docker pull amazon/dynamodb-local

# Run DynamoDB Local container
docker run -p 8000:8000 amazon/dynamodb-local
```

## Option 2: Manual Download

1. Download DynamoDB Local from AWS:
   https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.DownloadingAndRunning.html

2. Extract the archive to a directory of your choice

3. Navigate to the extracted directory and run:
   ```
   java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb
   ```

## Verifying DynamoDB Local Setup

Once DynamoDB Local is running, you can verify it's working by visiting:
http://localhost:8000/shell

## Using with Dev Reserve Server

Make sure your .env file contains:
```
DYNAMODB_ENDPOINT=http://localhost:8000
```

This tells the Go backend to use your local DynamoDB instead of the AWS service.
