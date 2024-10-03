#!/usr/bin/env python3

import sys
import click
from bs4 import BeautifulSoup, Comment
import re

def collapse_whitespace(text):
    # Replace multiple whitespace characters with a single space
    return re.sub(r'\s+', ' ', text.strip())

def simplify_html(soup, cleanup_whitespace, strip_pl_spans):
    # Remove comments
    for comment in soup.find_all(string=lambda text: isinstance(text, Comment)):
        comment.extract()

    # Remove script, style, meta, link, and noscript elements
    for tag in soup(['script', 'style', 'svg', 'meta', 'noscript', 'iframe', 'template']):
        tag.decompose()

    # Recursively process elements
    for tag in soup.find_all(True):
        # Keep only 'id' and 'class' attributes
        preserved_attrs = {}
        if 'id' in tag.attrs:
            preserved_attrs['id'] = tag.attrs['id']
        if 'class' in tag.attrs:
            preserved_attrs['class'] = tag.attrs['class']
        if 'src' in tag.attrs:
            preserved_attrs['src'] = tag.attrs['src']
        if 'href' in tag.attrs:
            preserved_attrs['href'] = tag.attrs['href']
        if 'alt' in tag.attrs:
            preserved_attrs['alt'] = tag.attrs['alt']
        tag.attrs = preserved_attrs

        # Strip pl-* spans if the flag is set
        if strip_pl_spans and tag.name == 'span' and tag.get('class'):
            if any(cls.startswith('pl-') for cls in tag.get('class')):
                tag.unwrap()

    # Apply whitespace cleanup if the flag is set
    if cleanup_whitespace:
        for text in soup.find_all(text=True):
            if text.parent.name not in ['script', 'style']:
                text.replace_with(text.strip())

    return soup

@click.command()
@click.option('-i', '--input', 'input_file', type=click.File('r'), default='-',
              help='Input HTML file (default: stdin)')
@click.option('-o', '--output', 'output_file', type=click.File('w'), default='-',
              help='Output HTML file (default: stdout)')
@click.option('--cleanup-whitespace', is_flag=True, default=False,
              help='Enable whitespace cleanup (default: False)')
@click.option('--strip-pl-spans', is_flag=True, default=True,
              help='Strip spans with pl-* classes used for code formatting (default: False)')
def main(input_file, output_file, cleanup_whitespace, strip_pl_spans):
    """Simplify HTML by removing unnecessary elements and attributes."""
    soup = BeautifulSoup(input_file, 'html.parser')
    simplified_soup = simplify_html(soup, cleanup_whitespace, strip_pl_spans)
    
    # Collapse whitespace on the final result if cleanup is enabled
    final_html = str(simplified_soup)
    if cleanup_whitespace:
        final_html = collapse_whitespace(final_html)
    
    output_file.write(final_html)

if __name__ == "__main__":
    main()
