import click
import yaml
from bs4 import BeautifulSoup
import sys

@click.command()
@click.option('--config', '-c', type=click.File('r'), help='YAML configuration file with selectors.')
def extract_content(config):
    """
    Extract content from HTML based on YAML-configured CSS selectors.
    """
    # Load YAML configuration
    config_data = yaml.safe_load(config)

    # Read HTML input from stdin
    html_input = sys.stdin.read()

    # Parse the HTML using BeautifulSoup
    soup = BeautifulSoup(html_input, 'html.parser')

    # Create a dictionary to hold the extracted data
    extracted_data = {}

    # Iterate through the selectors defined in the YAML configuration
    for item in config_data.get('selectors', []):
        title = item.get('title')
        selector = item.get('selector')
        
        # Extract the content using the CSS selector
        elements = soup.select(selector)
        
        # Store the extracted content in the dictionary
        extracted_data[title] = [element.get_text(strip=True) for element in elements]

    # Output the extracted data as YAML to stdout
    yaml_output = yaml.dump(extracted_data, default_flow_style=False, allow_unicode=True)
    sys.stdout.write(yaml_output)

if __name__ == '__main__':
    extract_content()
