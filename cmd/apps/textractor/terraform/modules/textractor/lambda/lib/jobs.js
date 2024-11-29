const path = require('path');

function sanitizeJobId(input) {
    // Remove file extension and replace invalid characters
    return input
        .replace(/\.[^/.]+$/, '') // remove extension
        .replace(/[^a-zA-Z0-9_.\-:]/g, '-'); // replace invalid chars with dash
}

function createJobId() {
    // Create a unique job ID using timestamp and random string
    return `job-${Date.now()}-${Math.random().toString(36).substring(2, 15)}`;
}

function getInputKey(jobId, fileName) {
    return `input/${jobId}/${fileName}`;
}

function getResultKey(textractJobId) {
    return `textract_output/${textractJobId}`;
}

function extractJobIdFromKey(key) {
    // Extract jobId from either input or results path
    const matches = key.match(/(?:input|results)\/([^\/]+)/);
    return matches ? matches[1] : null;
}

module.exports = {
    sanitizeJobId,
    createJobId,
    getInputKey,
    getResultKey,
    extractJobIdFromKey
}; 