const AWS = require('aws-sdk');
const { updateJobStatus, sendNotification } = require('./lib/db');
const textract = new AWS.Textract();
const s3 = new AWS.S3();

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
            console.error('Error processing completion:', err);
            throw err;
        }
    }
};