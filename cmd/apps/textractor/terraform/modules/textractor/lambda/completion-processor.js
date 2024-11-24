const AWS = require('aws-sdk');
const { updateJobStatus, sendNotification } = require('./lib/db');
const { getResultKey } = require('./lib/jobs');
const textract = new AWS.Textract();
const s3 = new AWS.S3();

// Add prefix to all console.log calls
const logPrefix = '[completion-processor]';

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
    const resultKey = getResultKey(jobId);
    
    await s3.putObject({
        Bucket: bucket,
        Key: resultKey,
        Body: JSON.stringify(results),
        ContentType: 'application/json'
    }).promise();

    return resultKey;
}

exports.handler = async (event) => {
    console.log(`${logPrefix} Processing SNS event:`, JSON.stringify(event));

    for (const record of event.Records) {
        try {
            const message = JSON.parse(record.body);
            const notification = JSON.parse(message.Message);

            const JobID = notification.JobTag;
            const textractJobId = notification.JobId;
            const status = notification.Status;

            console.log(`${logPrefix} Processing completion for job ${JobID} (Textract ID: ${textractJobId})`);

            if (status === 'SUCCEEDED') {
                const results = await getTextractResults(textractJobId);
                const resultKey = await saveResults(
                    process.env.STORAGE_BUCKET,
                    JobID,
                    results
                );

                const details = {
                    ResultKey: resultKey,
                    CompletedAt: new Date().toISOString()
                };

                await updateJobStatus(JobID, 'COMPLETED', details);
                await sendNotification(JobID, 'COMPLETED', details);
            } else {
                const details = {
                    Error: notification.StatusMessage || 'Textract processing failed',
                    CompletedAt: new Date().toISOString()
                };

                await updateJobStatus(JobID, 'FAILED', details);
                await sendNotification(JobID, 'FAILED', details);
            }
        } catch (err) {
            console.error(`${logPrefix} Error processing completion:`, err);
            // Try to extract JobID from the message if possible
            try {
                const message = JSON.parse(record.body);
                const notification = JSON.parse(message.Message);
                const JobID = notification.JobTag;
                
                const details = {
                    Error: err.message,
                    CompletedAt: new Date().toISOString()
                };
                
                await updateJobStatus(JobID, 'FAILED', details);
                await sendNotification(JobID, 'FAILED', details);
            } catch (innerErr) {
                console.error(`${logPrefix} Error while handling failure:`, innerErr);
            }
            // Don't rethrow - allow message to be deleted from queue
        }
    }
};