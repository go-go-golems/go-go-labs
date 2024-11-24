const AWS = require('aws-sdk');
const textract = new AWS.Textract();
const dynamodb = new AWS.DynamoDB.DocumentClient();
const s3 = new AWS.S3();
const sqs = new AWS.SQS();

const JOB_STATUS = {
    COMPLETED: 'COMPLETED',
    FAILED: 'ERROR'
};

exports.handler = async (event) => {
    console.log('Received Textract completion event:', JSON.stringify(event, null, 2));

    for (const record of event.Records) {
        try {
            const message = JSON.parse(record.Sns.Message);
            const textractJobId = message.JobId;
            const status = message.Status;
            
            const textractJob = await textract.getDocumentAnalysis({
                JobId: textractJobId
            }).promise();

            const jobId = textractJob.JobStatus.OutputConfig.S3Prefix.split('/')[1];

            if (status === 'SUCCEEDED') {
                let results = [];
                let nextToken = null;
                
                do {
                    const response = await textract.getDocumentAnalysis({
                        JobId: textractJobId,
                        NextToken: nextToken
                    }).promise();
                    results.push(response);
                    nextToken = response.NextToken;
                } while (nextToken);

                const resultKey = `results/${jobId}/analysis.json`;
                await s3.putObject({
                    Bucket: process.env.BUCKET_NAME,
                    Key: resultKey,
                    Body: JSON.stringify(results),
                    ContentType: 'application/json'
                }).promise();

                await updateJobCompletion(jobId, resultKey, textractJob.DocumentMetadata);
                await sendNotification({
                    jobId,
                    textractJobId,
                    status: JOB_STATUS.COMPLETED,
                    message: 'Document processing completed',
                    details: {
                        resultKey,
                        pages: results.reduce((acc, r) => acc + (r.Blocks?.length || 0), 0),
                        documentMetadata: textractJob.DocumentMetadata
                    }
                });
            } else {
                const errorMessage = textractJob.StatusMessage || 'Textract processing failed';
                await updateJobFailure(jobId, errorMessage);
                await sendNotification({
                    jobId,
                    textractJobId,
                    status: JOB_STATUS.FAILED,
                    message: errorMessage,
                    details: {
                        statusMessage: textractJob.StatusMessage,
                        documentMetadata: textractJob.DocumentMetadata
                    }
                });
            }
        } catch (err) {
            console.error('Error handling Textract completion:', err);
            if (err.jobId) {
                await sendNotification({
                    jobId: err.jobId,
                    status: JOB_STATUS.FAILED,
                    message: `System error: ${err.message}`
                });
            }
        }
    }
};

// ... helper functions ... 