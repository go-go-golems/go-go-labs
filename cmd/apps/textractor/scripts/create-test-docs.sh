#!/bin/bash

# Create directory for markdown sources
mkdir -p test-docs/markdown
mkdir -p test-docs/pdf

# Invoice example
cat > test-docs/markdown/invoice.md << 'EOL'
---
title: "Sample Invoice"
date: "2024-03-20"
---

# INVOICE

**Invoice #:** INV-2024-001  
**Date:** March 20, 2024  
**Due Date:** April 20, 2024  

## From
Acme Corporation  
123 Business St  
Business City, 12345  

## To
Client Company Ltd  
456 Client Avenue  
Client City, 67890  

| Item                     | Quantity | Rate    | Amount   |
|--------------------------|----------|---------|----------|
| Consulting Services      | 10       | $150.00 | $1500.00 |
| Software Development     | 20       | $200.00 | $4000.00 |
| Project Management       | 5        | $175.00 | $875.00  |

**Subtotal:** $6375.00  
**Tax (10%):** $637.50  
**Total:** $7012.50  
EOL

# Text with tables
cat > test-docs/markdown/tables.md << 'EOL'
---
title: "Sample Report with Tables"
date: "2024-03-20"
---

# Quarterly Sales Report

## Regional Performance

| Region    | Q1 Sales | Q2 Sales | Q3 Sales | Q4 Sales |
|-----------|----------|----------|----------|----------|
| North     | $50,000  | $65,000  | $70,000  | $85,000  |
| South     | $45,000  | $48,000  | $52,000  | $58,000  |
| East      | $62,000  | $67,000  | $71,000  | $76,000  |
| West      | $55,000  | $59,000  | $63,000  | $68,000  |

## Product Categories

| Category      | Units Sold | Revenue   | Profit   |
|--------------|------------|-----------|----------|
| Electronics   | 1,200      | $360,000  | $108,000 |
| Furniture     | 800        | $240,000  | $72,000  |
| Office Supply | 2,500      | $125,000  | $37,500  |
EOL

# Text with forms
cat > test-docs/markdown/forms.md << 'EOL'
---
title: "Employment Application Form"
date: "2024-03-20"
---

# Employment Application

## Personal Information

$\square$ Mr. $\square$ Mrs. $\square$ Ms. $\square$ Dr.

**Full Name:** ________________________________________________

**Address:** _________________________________________________

**Phone:** ___________________________________________________

**Email:** ___________________________________________________

## Education

**Highest Degree:** $\square$ High School $\square$ Bachelor's $\square$ Master's $\square$ PhD

**Institution:** ______________________________________________

**Year of Graduation:** ______________________________________

## Employment History

**Current Employer:** ________________________________________

**Position:** _______________________________________________

**Years of Experience:** ____________________________________

$\square$ I certify that all information provided is accurate
$\square$ I agree to background verification

Signature: ___________________ Date: ____________________
EOL

# Regular text document
cat > test-docs/markdown/text.md << 'EOL'
---
title: "Sample Text Document"
date: "2024-03-20"
---

# Project Proposal

## Executive Summary

This document outlines the proposed implementation of a new customer relationship management system. The project aims to streamline our customer service operations and improve data analytics capabilities.

## Background

Our current system has been in place for five years and lacks modern features necessary for efficient operation. Customer data is scattered across multiple platforms, making it difficult to maintain a coherent view of customer interactions.

## Objectives

1. Consolidate customer data into a single platform
2. Improve response time to customer inquiries
3. Enable advanced analytics and reporting
4. Reduce operational costs

## Implementation Timeline

The project will be implemented in phases over six months, with regular checkpoints and reviews to ensure alignment with business objectives.

## Budget Considerations

The estimated budget for this project includes software licensing, implementation costs, and staff training. A detailed breakdown will be provided in the following section.
EOL

# Convert all markdown files to PDF using pandoc
for file in test-docs/markdown/*.md; do
    filename=$(basename "$file" .md)
    echo "Converting $filename.md to PDF..."
    pandoc "$file" \
        -f markdown \
        -t pdf \
        --pdf-engine=pdflatex \
        -o "test-docs/pdf/$filename.pdf"
done

echo "Test documents have been created in test-docs/pdf/" 