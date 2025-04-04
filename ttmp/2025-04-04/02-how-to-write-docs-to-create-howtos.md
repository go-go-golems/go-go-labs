# Prompt for Creating Technical How-To Documents

## Overview

This prompt is designed to help Large Language Models (LLMs) generate comprehensive, well-structured technical how-to documents similar to the example provided about building a talks scheduling application.

## The Prompt

```
Create a comprehensive technical how-to document for [SPECIFIC TASK/APPLICATION]. 

Your document should follow this structure:

1. Introduction
   - Explain the purpose and value of the application/system
   - Identify the target audience and their needs
   - Provide a brief overview of what will be covered

2. Requirements and Specifications
   - List functional requirements (what the system needs to do)
   - Include any technical constraints or preferences
   - Prioritize features (core vs. nice-to-have)

3. Architecture Overview
   - Describe the high-level architecture using a Mermaid diagram
   - Identify major components and how they interact
   - Explain technology choices with rationales

4. Data Model
   - Define key entities and their relationships
   - Detail important fields/attributes for each entity
   - Explain any special data handling considerations

5. Implementation Plan
   - Break down the development into logical phases
   - Provide directory/file structure recommendations
   - Include code snippets for core functionality (10-20 lines each)

6. Key Features Implementation
   - Provide detailed implementation approaches for 2-3 critical features
   - Include pseudocode or actual code examples
   - Highlight potential challenges and solutions

7. UI/UX Considerations
   - Describe main interface elements and user flows
   - Explain design principles being followed
   - List key screens/views with their purpose

8. Deployment Strategy
   - Outline deployment options with pros/cons
   - Include infrastructure considerations
   - Mention any CI/CD recommendations

9. Future Enhancements
   - Suggest 3-5 potential future improvements
   - Briefly explain the value of each enhancement

10. Conclusion
    - Summarize the approach and its benefits
    - Reinforce how this solution addresses the original needs

For all code examples:
- Use clean, well-commented code that follows best practices for [LANGUAGE/FRAMEWORK]
- Include error handling when appropriate
- Follow standard naming conventions

For diagrams:
- Use Mermaid diagram syntax for architecture, data flow, and entity relationships
- Include clear labels and descriptions for each component

Make the document practical and actionable, with sufficient detail that a developer could begin implementation, while keeping explanations concise and focused.
```

## Guidelines for Effective Results

When using this prompt with an LLM, consider these additional guidelines to get optimal results:

1. **Be specific about technologies**: Replace `[LANGUAGE/FRAMEWORK]` with specific technologies (e.g., "Go with HTMX and Bootstrap").

2. **Provide context**: Include any relevant background about the project's constraints, team size, or special requirements.

3. **Request visual elements**: Ask for Mermaid diagrams for architecture, data models, and user flows.

4. **Encourage pragmatism**: Ask the LLM to focus on practical, implementable solutions rather than theoretical ideals.

5. **Request progressive disclosure**: Have complex topics explained in layers, with high-level overviews followed by details.

6. **Guide the tone**: Specify whether you want a more formal technical document or a conversational tutorial style.

7. **Set scope boundaries**: Clearly define what aspects of the system should be covered in depth vs. briefly mentioned.

## Example Modifications

For a frontend-focused document:
```
Emphasize the UI/UX section with wireframes described in detail, and include specific component structures for [FRAMEWORK].
```

For a backend-focused document:
```
Expand the data model and API design sections. Include database schema details, API endpoints, and authentication strategies.
```

For a DevOps-focused document:
```
Prioritize deployment, scaling, monitoring, and maintenance aspects. Include configuration examples for [CLOUD PROVIDER/PLATFORM].
```

## Tips for Reviewing the Generated Document

After the LLM generates the document:

1. Check for technical accuracy and feasibility
2. Ensure all major components are addressed
3. Verify that code examples are complete and correct
4. Look for logical gaps in the implementation plan
5. Confirm that the architecture matches the requirements
6. Ensure Mermaid diagrams are correctly formatted and meaningful

The goal is to produce a document that balances comprehensive guidance with practical implementation details, serving as both a roadmap and a reference for development. 