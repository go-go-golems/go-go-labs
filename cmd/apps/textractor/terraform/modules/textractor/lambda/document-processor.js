const AWS = require('aws-sdk');
const { updateJobStatus, sendNotification } = require('./lib/db');
const { sanitizeJobId, extractJobIdFromKey } = require('./lib/jobs');
const textract = new AWS.Textract();

// Add prefix to all console.log calls
const logPrefix = '[document-processor]';

async function startTextractJob(bucket, key, jobId) {
    const params = {
        DocumentLocation: {
            S3Object: {
                Bucket: bucket,
                Name: key
            }
        },
        JobTag: sanitizeJobId(jobId),
        NotificationChannel: {
            RoleArn: process.env.TEXTRACT_ROLE_ARN,
            SNSTopicArn: process.env.SNS_TOPIC_ARN
        },
        FeatureTypes: ['TABLES', 'FORMS']
    };

    // Add ClientRequestToken for idempotency
    if (process.env.CLIENT_REQUEST_TOKEN) {
        params.ClientRequestToken = process.env.CLIENT_REQUEST_TOKEN;
    }

    console.log(`${logPrefix} Starting Textract job with params:`, JSON.stringify(params, null, 2));
    try {
        const response = await textract.startDocumentAnalysis(params).promise();
        console.log(`${logPrefix} Textract job started successfully for ${bucket}/${key}. Job ID: ${response.JobId}`);
        return response.JobId;
    } catch (error) {
        console.error(`${logPrefix} Error starting Textract job for ${bucket}/${key}:`, JSON.stringify(error, null, 2));
        throw error;
    }
}

exports.handler = async (event) => {
    console.log(`${logPrefix} Received event:`, JSON.stringify(event, null, 2));

    for (const record of event.Records) {
        try {
            const s3Event = JSON.parse(record.body);

            for (const s3Record of s3Event.Records) {
                const bucket = s3Record.s3.bucket.name;
                const key = decodeURIComponent(s3Record.s3.object.key.replace(/\+/g, ' '));
                
                console.log(`${logPrefix} Processing new file: ${bucket}/${key}`);

                const jobId = extractJobIdFromKey(key);
                if (!jobId) {
                    throw new Error(`Could not extract jobId from key: ${key}`);
                }

                await updateJobStatus(jobId, 'PROCESSING');

                console.log(`${logPrefix} Starting Textract job for ${bucket}/${key}`);
                const textractJobId = await startTextractJob(bucket, key, jobId);
                await updateJobStatus(jobId, 'PROCESSING', { textractJobId });
            }
        } catch (err) {
            console.error(`${logPrefix} Error processing record:`, err);
            // Update job status to failed and send notification
            const jobId = JSON.parse(record.body).Records[0].s3.object.key.replace(/\.[^/.]+$/, '');
            await updateJobStatus(jobId, 'FAILED', { error: err.message });
            await sendNotification(jobId, 'FAILED', { error: err.message });
            // Don't rethrow the error - this will allow SQS to remove the message
            // from the queue as it's been "successfully" processed
        }
    }
};