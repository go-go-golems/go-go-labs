const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();
const textract = new AWS.Textract();
const sns = new AWS.SNS();

// Update job status in DynamoDB
async function updateJobStatus(JobID, status, additionalData = {}) {
    const params = {
        TableName: process.env.JOBS_TABLE,
        Key: {
            JobID: JobID
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

    // Add any additional data to the update
    if (Object.keys(additionalData).length > 0) {
        params.UpdateExpression += ', ';
        Object.entries(additionalData).forEach(([key, value], index) => {
            params.UpdateExpression += `#add${index} = :add${index}${index < Object.keys(additionalData).length - 1 ? ', ' : ''}`;
            params.ExpressionAttributeNames[`#add${index}`] = key;
            params.ExpressionAttributeValues[`:add${index}`] = value;
        });
    }

    try {
        await dynamodb.update(params).promise();
    } catch (err) {
        console.error('Failed to update job status:', err);
        throw err;
    }
}

// Start Textract processing
async function startTextractJob(bucket, key, jobId) {
    const params = {
        DocumentLocation: {
            S3Object: {
                Bucket: bucket,
                Name: key
            }
        },
        JobTag: jobId,
        NotificationChannel: {
            RoleArn: process.env.TEXTRACT_ROLE_ARN,
            SNSTopicArn: process.env.SNS_TOPIC_ARN
        },
        FeatureTypes: ['TABLES', 'FORMS']
    };

    const response = await textract.startDocumentAnalysis(params).promise();
    return response.JobId;
}

exports.handler = async (event) => {
    console.log('Received event:', JSON.stringify(event, null, 2));

    for (const record of event.Records) {
        try {
            const s3Event = JSON.parse(record.body);

            for (const s3Record of s3Event.Records) {
                const bucket = s3Record.s3.bucket.name;
                const key = decodeURIComponent(s3Record.s3.object.key.replace(/\+/g, ' '));
                
                console.log(`Processing new file: ${bucket}/${key}`);

                // Use the full input path as the job ID to ensure uniqueness
                const jobId = key.replace(/\.[^/.]+$/, ''); // Removes file extension but keeps full path

                // Update initial job status
                await updateJobStatus(jobId, 'PROCESSING');

                // Start Textract job
                const textractJobId = await startTextractJob(bucket, key, jobId);

                // Update job with Textract ID
                await updateJobStatus(jobId, 'PROCESSING', { textractJobId });
            }
        } catch (err) {
            console.error('Error processing record:', err);
            throw err;
        }
    }
};