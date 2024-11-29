#!/bin/bash

# Array of all Textract API objects
OBJECTS=(
    "Adapter"
    "AdapterOverview"
    "AdaptersConfig"
    "AdapterVersionDatasetConfig"
    "AdapterVersionEvaluationMetric"
    "AdapterVersionOverview"
    "AnalyzeIDDetections"
    "Block"
    "BoundingBox"
    "DetectedSignature"
    "Document"
    "DocumentGroup"
    "DocumentLocation"
    "DocumentMetadata"
    "EvaluationMetric"
    "ExpenseCurrency"
    "ExpenseDetection"
    "ExpenseDocument"
    "ExpenseField"
    "ExpenseGroupProperty"
    "ExpenseType"
    "Extraction"
    "Geometry"
    "HumanLoopActivationOutput"
    "HumanLoopConfig"
    "HumanLoopDataAttributes"
    "IdentityDocument"
    "IdentityDocumentField"
    "LendingDetection"
    "LendingDocument"
    "LendingField"
    "LendingResult"
    "LendingSummary"
    "LineItemFields"
    "LineItemGroup"
    "NormalizedValue"
    "NotificationChannel"
    "OutputConfig"
    "PageClassification"
    "Point"
    "Prediction"
    "QueriesConfig"
    "Query"
    "Relationship"
    "S3Object"
    "SignatureDetection"
    "SplitDocument"
    "UndetectedSignature"
    "Warning"
)

# Create a directory for the downloaded files
mkdir -p textract-api-docs

# Download documentation for each object
for object in "${OBJECTS[@]}"; do
    echo "Downloading documentation for ${object}..."
    
    # Create the URL
    url="https://docs.aws.amazon.com/en_us/textract/latest/dg/API_${object}.html"
    output_file="textract-api-docs/api-${object}.txt"
    
    # Download and convert to text
    if lynx -nolist -dump "$url" > "$output_file"; then
        echo "Successfully downloaded documentation for ${object}"
    else
        echo "Error downloading documentation for ${object}"
    fi
    
    # Add a small delay to avoid overwhelming the server
    sleep 1
done

echo "Download complete. Documentation files are saved in the textract-api-docs directory."
