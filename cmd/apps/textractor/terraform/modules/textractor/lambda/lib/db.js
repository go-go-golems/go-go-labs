const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();
const sns = new AWS.SNS();

// Add prefix to all console.log calls
const logPrefix = '[db-lib]';

async function updateJobStatus(JobID, status, additionalData = {}) {
    const params = {
        TableName: process.env.JOBS_TABLE,
        Key: {
            JobID
        },
        UpdateExpression: 'SET #status = :status, UpdatedAt = :updatedAt',
        ExpressionAttributeNames: {
            '#status': 'Status'
        },
        ExpressionAttributeValues: {
            ':status': status,
            ':updatedAt': new Date().toISOString()
        }
    };

    // Add additional data to update expression if provided
    if (Object.keys(additionalData).length > 0) {
        Object.entries(additionalData).forEach(([key, value]) => {
            params.UpdateExpression += `, ${key} = :${key}`;
            params.ExpressionAttributeValues[`:${key}`] = value;
        });
    }

    console.log(`[db-lib] Updating job status with params:`, JSON.stringify(params, null, 2));
    
    try {
        await dynamodb.update(params).promise();
        console.log(`[db-lib] Successfully updated job status for ${JobID} to ${status}`);
    } catch (error) {
        console.error(`[db-lib] Error updating job status:`, error);
        throw error;
    }
}

exports.updateJobStatus = updateJobStatus;

exports.sendNotification = async (jobId, status, details = {}) => {
    const message = {
        jobId,
        status,
        ...details,
        timestamp: new Date().toISOString()
    };

    console.log(`${logPrefix} Sending notification for ${jobId} with status ${status}`);
    console.log(`${logPrefix} Full notification payload:`, JSON.stringify(message, null, 2));
    await sns.publish({
        TopicArn: process.env.NOTIFICATION_TOPIC_ARN,
        Message: JSON.stringify(message)
    }).promise();
}; 