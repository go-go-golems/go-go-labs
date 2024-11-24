const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();
const textract = new AWS.Textract();
const s3 = new AWS.S3();
const sns = new AWS.SNS();

async function updateJobStatus(JobID, status, details = {}) {
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
}

async function getTextractResults(textractJobId) {
    const results = [];
    let nextToken = null;

    do {
        const params = {
            JobId: textractJobId,
            MaxResults: 1000,
            NextToken: nextToken
        };

        const response = await textract.getDocumentAnalysis(params).promise();
        results.push(...response.Blocks);
        nextToken = response.NextToken;
    } while (nextToken);

    return results;
}

async function saveResults(bucket, jobId, results) {
    const resultKey = `results/${jobId}/analysis.json`;
    
    await s3.putObject({
        Bucket: bucket,
        Key: resultKey,
        Body: JSON.stringify(results),
        ContentType: 'application/json'
    }).promise();

    return resultKey;
}

async function sendNotification(jobId, status, details = {}) {
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
}

exports.handler = async (event) => {
    console.log('Received event:', JSON.stringify(event, null, 2));

    for (const record of event.Records) {
        try {
            const message = JSON.parse(record.body);
            const notification = JSON.parse(message.Message);

            const JobID = notification.JobTag;
            const textractJobId = notification.JobId;
            const status = notification.Status;

            console.log(`Processing completion for job ${JobID} (Textract ID: ${textractJobId})`);

            if (status === 'SUCCEEDED') {
                // Get results from Textract
                const results = await getTextractResults(textractJobId);

                // Save results to S3
                const resultKey = await saveResults(
                    process.env.STORAGE_BUCKET,
                    JobID,
                    results
                );

                const details = {
                    ResultKey: resultKey,
                    CompletedAt: new Date().toISOString()
                };

                // Update job status
                await updateJobStatus(JobID, 'COMPLETED', details);
                
                // Send completion notification
                await sendNotification(JobID, 'COMPLETED', details);
            } else {
                const details = {
                    Error: notification.StatusMessage || 'Textract processing failed',
                    CompletedAt: new Date().toISOString()
                };

                // Handle failure
                await updateJobStatus(JobID, 'FAILED', details);
                
                // Send failure notification
                await sendNotification(JobID, 'FAILED', details);
            }
        } catch (err) {
            console.error('Error processing completion:', err);
            throw err;
        }
    }
};