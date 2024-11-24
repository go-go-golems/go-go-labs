const AWS = require('aws-sdk');
const textract = new AWS.Textract();
const sns = new AWS.SNS();
const dynamodb = new AWS.DynamoDB.DocumentClient();
const s3 = new AWS.S3();
const sqs = new AWS.SQS();

// Job states
const JOB_STATUS = {
    PROCESSING: 'PROCESSING',
    COMPLETED: 'COMPLETED',
    FAILED: 'ERROR'
};

exports.handler = async (event) => {
    console.log('Received event:', JSON.stringify(event, null, 2));
    
    // Check if this is an SNS notification (Textract completion)
    if (event.Records?.[0]?.Sns) {
        return await handleTextractCompletion(event);
    }
    
    // Process SQS messages (new document uploads)
    for (const record of event.Records) {
        try {
            // Parse the message body
            const body = JSON.parse(record.body);
            console.log('Processing message:', JSON.stringify(body, null, 2));
            
            // If this is an S3 event
            if (body.Records) {
                for (const s3Record of body.Records) {
                    if (s3Record.eventName.startsWith('ObjectCreated:')) {
                        const bucket = s3Record.s3.bucket.name;
                        const key = decodeURIComponent(s3Record.s3.object.key.replace(/\+/g, ' '));
                        
                        console.log(`Processing new file: ${bucket}/${key}`);
                        
                        // Extract jobId from the key (input/<jobId>/filename.pdf)
                        const jobId = key.split('/')[1];
                        if (!jobId) {
                            console.error('Could not extract jobId from key:', key);
                            continue;
                        }

                        try {
                            // Update job status to PROCESSING
                            await updateJobStatus(jobId, JOB_STATUS.PROCESSING);

                            // Start Textract job
                            const textractJobId = await startTextractJob(bucket, key);
                            console.log(`Started Textract job ${textractJobId} for document ${key}`);

                            // Update job with Textract ID
                            await updateTextractJobId(jobId, textractJobId);

                            // Send SNS notification
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
            // Don't throw the error - we want to continue processing other records
        }
    }
    
    return {
        statusCode: 200,
        body: JSON.stringify('Processing complete')
    };
};

async function startTextractJob(bucket, key) {
    const params = {
        DocumentLocation: {
            S3Object: {
                Bucket: bucket,
                Name: key
            }
        },
        FeatureTypes: ['TABLES', 'FORMS', 'SIGNATURES'],
        NotificationChannel: {
            RoleArn: process.env.AWS_LAMBDA_ROLE,
            SNSTopicArn: process.env.SNS_TOPIC_ARN
        },
        OutputConfig: {
            S3Bucket: bucket,
            S3Prefix: `results/${key.split('/')[1]}/` // Use jobId as prefix
        }
    };

    const response = await textract.startDocumentAnalysis(params).promise();
    return response.JobId;
}

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

    if (status === JOB_STATUS.COMPLETED || status === JOB_STATUS.FAILED) {
        params.UpdateExpression += ', CompletedAt = :completedAt';
        params.ExpressionAttributeValues[':completedAt'] = new Date().toISOString();
    }
    
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

async function sendNotification(message) {
    // Send to SNS
    const snsParams = {
        Message: JSON.stringify(message),
        TopicArn: process.env.SNS_TOPIC_ARN
    };
    await sns.publish(snsParams).promise();

    // Send to notifications queue
    const sqsParams = {
        MessageBody: JSON.stringify({
            type: 'STATUS_UPDATE',
            timestamp: new Date().toISOString(),
            ...message
        }),
        QueueUrl: process.env.NOTIFICATIONS_QUEUE_URL,
        MessageAttributes: {
            jobId: {
                DataType: 'String',
                StringValue: message.jobId
            }
        }
    };
    await sqs.sendMessage(sqsParams).promise();
}

// Handle Textract completion notifications
exports.handleTextractCompletion = async (event) => {
    console.log('Received Textract completion event:', JSON.stringify(event, null, 2));

    for (const record of event.Records) {
        try {
            const message = JSON.parse(record.Sns.Message);
            const textractJobId = message.JobId;
            const status = message.Status;
            
            // Get job details from Textract
            const textractJob = await textract.getDocumentAnalysis({
                JobId: textractJobId
            }).promise();

            // Extract jobId from the OutputConfig S3Prefix
            const jobId = textractJob.JobStatus.OutputConfig.S3Prefix.split('/')[1];

            if (status === 'SUCCEEDED') {
                // Get all results pages
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

                // Store results in S3
                const resultKey = `results/${jobId}/analysis.json`;
                await s3.putObject({
                    Bucket: process.env.BUCKET_NAME,
                    Key: resultKey,
                    Body: JSON.stringify(results),
                    ContentType: 'application/json'
                }).promise();

                // Update job status and result location
                await dynamodb.update({
                    TableName: process.env.JOBS_TABLE,
                    Key: { JobID: jobId },
                    UpdateExpression: 'SET #status = :status, ResultKey = :rk, CompletedAt = :cat',
                    ExpressionAttributeNames: {
                        '#status': 'Status'
                    },
                    ExpressionAttributeValues: {
                        ':status': JOB_STATUS.COMPLETED,
                        ':rk': resultKey,
                        ':cat': new Date().toISOString()
                    }
                }).promise();

                // Send detailed completion notification
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
                await updateJobStatus(jobId, JOB_STATUS.FAILED, errorMessage);
                
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
            // Try to send error notification if possible
            if (err.jobId) {
                await sendNotification({
                    jobId: err.jobId,
                    status: JOB_STATUS.ERROR,
                    message: `System error: ${err.message}`
                });
            }
        }
    }
}; 