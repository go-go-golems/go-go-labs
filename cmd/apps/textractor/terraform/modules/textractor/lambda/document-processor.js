const AWS = require('aws-sdk');
const textract = new AWS.Textract();
const sns = new AWS.SNS();
const dynamodb = new AWS.DynamoDB.DocumentClient();
const s3 = new AWS.S3();

const JOB_STATUS = {
    PROCESSING: 'PROCESSING',
    FAILED: 'ERROR'
};

exports.handler = async (event) => {
    console.log('Received event:', JSON.stringify(event, null, 2));
    
    for (const record of event.Records) {
        try {
            const body = JSON.parse(record.body);
            console.log('Processing message:', JSON.stringify(body, null, 2));
            
            if (body.Records) {
                for (const s3Record of body.Records) {
                    if (s3Record.eventName.startsWith('ObjectCreated:')) {
                        const bucket = s3Record.s3.bucket.name;
                        const key = decodeURIComponent(s3Record.s3.object.key.replace(/\+/g, ' '));
                        
                        console.log(`Processing new file: ${bucket}/${key}`);
                        
                        const jobId = key.split('/')[1];
                        if (!jobId) {
                            console.error('Could not extract jobId from key:', key);
                            continue;
                        }

                        try {
                            await updateJobStatus(jobId, JOB_STATUS.PROCESSING);
                            const textractJobId = await startTextractJob(bucket, key);
                            console.log(`Started Textract job ${textractJobId} for document ${key}`);
                            await updateTextractJobId(jobId, textractJobId);
                            await sendNotification({
                                jobId,
                                textractJobId,
                                status: JOB_STATUS.PROCESSING,
                                message: 'Document processing started'
                            });
                        } catch (err) {
                            console.error('Error processing document:', err);
                            await updateJobStatus(jobId, JOB_STATUS.FAILED, err.message);
                            await sendNotification({
                                jobId,
                                status: JOB_STATUS.FAILED,
                                message: err.message
                            });
                        }
                    }
                }
            }
        } catch (err) {
            console.error('Error processing record:', err);
        }
    }
    
    return { statusCode: 200, body: JSON.stringify('Processing complete') };
};

// ... helper functions (startTextractJob, updateJobStatus, etc.) ... 