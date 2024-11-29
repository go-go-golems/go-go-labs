# Textractor Command Implementation Plan

## Bugs
- [ ] Textract ID or Result Key do not seem to be populated
- [ ] Print out s3 path in list
- [ ] Fetch can use job-id

## Core Infrastructure (Required First)
- [x] DynamoDB table setup
- [x] Basic CLI structure
- [x] Resource loading from Terraform
- [x] Job tracking model

## submit
- [x] Generate unique JobID (UUID v4)
- [x] Support single file upload
- [x] Support directory upload (recursive)
- [x] Create DynamoDB job entry (UPLOADING status)
- [x] Upload PDF to S3 with content-type
- [x] Handle upload errors gracefully (ERROR status)
- [x] Update job status after successful upload (SUBMITTED)
- [ ] *Optional: Watch directory for new files*
- [ ] *Optional: Progress bar for large files*
- [ ] *Optional: Batch upload optimization*

## status
- [ ] Query job by ID from DynamoDB
- [ ] Display detailed job information
- [ ] Show processing progress if available
- [ ] Include error details if failed
- [ ] *Optional: Status history tracking*
- [ ] *Optional: Estimated completion time*

## fetch
- [ ] Retrieve Textract results from S3
- [ ] Support multiple output formats (JSON, Text)
- [ ] Handle partial results
- [ ] Support output file specification
- [ ] *Optional: Streaming large results*
- [ ] *Optional: Resume interrupted downloads*

## list
- [x] Query jobs from DynamoDB
- [x] Filter by status
- [x] Filter by date range
- [ ] Pagination support
- [ ] *Optional: Advanced filtering (size, type)*
- [ ] *Optional: Custom output formats*

## purge
- [ ] Delete job records from DynamoDB
- [ ] Remove associated S3 objects
- [ ] Support batch deletion
- [ ] Age-based cleanup
- [ ] Status-based cleanup
- [ ] *Optional: Soft delete with recovery window*
- [ ] *Optional: Audit log of deletions*

## monitor
- [ ] Real-time status updates
- [ ] DynamoDB stream integration
- [ ] Console-based UI
- [ ] Error notification
- [ ] Rate/cost monitoring
- [ ] *Optional: Webhook notifications*
- [ ] *Optional: Slack/Discord integration*
- [ ] *Optional: Custom alert thresholds*

## export
- [ ] Text extraction
- [ ] Table extraction
- [ ] Form field extraction
- [ ] Support multiple formats:
  - [ ] Plain text
  - [ ] Markdown tables
  - [ ] CSV for tables
  - [ ] JSON structure
- [ ] *Optional: PDF annotation overlay*
- [ ] *Optional: HTML output with styling*

## analyze
- [ ] Table structure analysis
- [ ] Form field detection
- [ ] Key-value pair extraction
- [ ] Signature detection
- [ ] *Optional: Custom field extraction*
- [ ] *Optional: Template matching*
- [ ] *Optional: Data validation rules*

## estimate
- [ ] Calculate page count
- [ ] Estimate AWS costs:
  - [ ] Textract processing
  - [ ] S3 storage
  - [ ] DynamoDB usage
- [ ] Batch estimation
- [ ] *Optional: Historical cost analysis*
- [ ] *Optional: Cost optimization suggestions*

## Shared Components
- [ ] AWS session management
- [ ] Error handling framework
- [ ] Logging infrastructure
- [ ] Rate limiting
- [ ] Retry logic
- [ ] Progress reporting
- [ ] *Optional: Telemetry/metrics*

## Testing Requirements
- [ ] Unit tests for each command
- [ ] Integration tests with AWS
- [ ] Sample PDF test suite
- [ ] Performance benchmarks
- [ ] *Optional: Chaos testing*

## Documentation
- [ ] Command reference
- [ ] Example workflows
- [ ] Troubleshooting guide
- [ ] AWS cost guidance
- [ ] *Optional: Video tutorials*

Note: Items marked with * are enhancements that add value but aren't critical for initial release. Implementation order should prioritize core functionality first. 