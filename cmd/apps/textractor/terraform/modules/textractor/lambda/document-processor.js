const AWS = require('aws-sdk');
const { updateJobStatus } = require('./lib/db');
const textract = new AWS.Textract();

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

                const jobId = key.replace(/\.[^/.]+$/, '');
                await updateJobStatus(jobId, 'PROCESSING');

                const textractJobId = await startTextractJob(bucket, key, jobId);
                await updateJobStatus(jobId, 'PROCESSING', { textractJobId });
            }
        } catch (err) {
            console.error('Error processing record:', err);
            throw err;
        }
    }
};