#!/usr/bin/env python3

import re
import sys

def clean_span_code_block(match):
    html_text = match.group(1)
    
    # Remove all HTML span tags
    text = re.sub(r'</?span[^>]*>', '', html_text)
    
    # Fix indentation and line breaks
    lines = text.split('\n')
    cleaned_lines = []
    for line in lines:
        # Remove leading/trailing whitespace
        cleaned_line = line.strip()
        if cleaned_line:
            cleaned_lines.append(cleaned_line)
    
    # Join with proper line breaks for code
    code = '\n'.join(cleaned_lines)
    
    # Format as markdown code block
    return f"```go\n{code}\n```"

async def process_file(filename):
    try:
        with open(filename, 'r', encoding='utf-8') as file:
            content = file.read()
        
        # Find code blocks with span tags and replace them
        pattern = r'```\s*\n((?:<span>[\s\S]*?</span>[\s\S]*?)+)\n```'
        cleaned_content = re.sub(pattern, clean_span_code_block, content)
        
        # Output the cleaned content
        print(cleaned_content)
        
        # Optionally save to a new file
        # with open(f"{filename}.cleaned.md", 'w', encoding='utf-8') as file:
        #     file.write(cleaned_content)
        
    except FileNotFoundError:
        print(f"Error: File '{filename}' not found")
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

async def main():
    if len(sys.argv) < 2:
        print("Usage: python script.py <filename>")
        sys.exit(1)
    
    await process_file(sys.argv[1])

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
