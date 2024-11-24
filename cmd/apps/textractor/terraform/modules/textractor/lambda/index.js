const AWS = require('aws-sdk');
const textract = new AWS.Textract();
const sns = new AWS.SNS();

exports.handler = async (event) => {
    const s3Record = event.Records[0].s3;
    const bucket = s3Record.bucket.name;
    const key = decodeURIComponent(s3Record.object.key.replace(/\+/g, ' '));
    
    try {
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
        
        return {
            statusCode: 200,
            body: JSON.stringify({
                message: 'Document analysis started',
                jobId: startResponse.JobId
            })
        };
    } catch (error) {
        console.error('Error:', error);
        throw error;
    }
}; 