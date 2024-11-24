const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();
const sns = new AWS.SNS();

exports.updateJobStatus = async (JobID, status, details = {}) => {
    const params = {
        TableName: process.env.JOBS_TABLE,
        Key: { JobID },
        UpdateExpression: 'SET #status = :status, UpdatedAt = :now',
        ExpressionAttributeNames: {
            '#status': 'Status'
        },
        ExpressionAttributeValues: {
            ':status': status,
            ':now': new Date().toISOString()
        }
    };

    if (Object.keys(details).length > 0) {
        params.UpdateExpression += ', Details = :details';
        params.ExpressionAttributeValues[':details'] = details;
    }

    await dynamodb.update(params).promise();
};

exports.sendNotification = async (jobId, status, details = {}) => {
    const message = {
        jobId,
        status,
        ...details,
        timestamp: new Date().toISOString()
    };

    await sns.publish({
        TopicArn: process.env.NOTIFICATION_TOPIC_ARN,
        Message: JSON.stringify(message)
    }).promise();
}; 