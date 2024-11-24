const AWS = require('aws-sdk');
const textract = new AWS.Textract();
const sns = new AWS.SNS();
const dynamodb = new AWS.DynamoDB.DocumentClient();

exports.handler = async (event) => {
    const s3Record = event.Records[0].s3;
    const bucket = s3Record.bucket.name;
    const key = decodeURIComponent(s3Record.object.key.replace(/\+/g, ' '));
    
    // Extract jobId from the key (assuming format: input/{jobId}/filename.pdf)
    const jobId = key.split('/')[1];
    
    try {
        // Update job status to PROCESSING
        await updateJobStatus(jobId, 'PROCESSING');
        
        // Start document analysis
        const params = {
            DocumentLocation: {
                S3Object: {
                    Bucket: bucket,
                    Name: key
                }
            },
            FeatureTypes: ['FORMS', 'TABLES'],
            NotificationChannel: {
                SNSTopicArn: process.env.SNS_TOPIC_ARN,
                RoleArn: process.env.AWS_LAMBDA_ROLE
            }
        };
        
        const startResponse = await textract.startDocumentAnalysis(params).promise();
        
        // Update job with Textract JobId
        await updateTextractJobId(jobId, startResponse.JobId);
        
        return {
            statusCode: 200,
            body: JSON.stringify({
                message: 'Document analysis started',
                jobId: startResponse.JobId
            })
        };
    } catch (error) {
        console.error('Error:', error);
        await updateJobStatus(jobId, 'ERROR', error.message);
        throw error;
    }
};

async function updateJobStatus(jobId, status, errorMsg = '') {
    const params = {
        TableName: process.env.JOBS_TABLE,
        Key: { JobID: jobId },
        UpdateExpression: 'SET #status = :status, #error = :error',
        ExpressionAttributeNames: {
            '#status': 'Status',
            '#error': 'Error'
        },
        ExpressionAttributeValues: {
            ':status': status,
            ':error': errorMsg
        }
    };
    
    await dynamodb.update(params).promise();
}

async function updateTextractJobId(jobId, textractJobId) {
    const params = {
        TableName: process.env.JOBS_TABLE,
        Key: { JobID: jobId },
        UpdateExpression: 'SET TextractID = :tid',
        ExpressionAttributeValues: {
            ':tid': textractJobId
        }
    };
    
    await dynamodb.update(params).promise();
} 