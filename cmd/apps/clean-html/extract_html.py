import click
import yaml
from bs4 import BeautifulSoup
import sys

def apply_transformations(text, transformations):
    for transform in transformations:
        if transform == "strip":
            text = text.strip()
        elif transform == "capitalize":
            text = text.capitalize()
        elif transform == "remove_newlines":
            text = text.replace('\n', '')
        elif transform == "to_lowercase":
            text = text.lower()
        elif transform == "to_uppercase":
            text = text.upper()
        elif transform == "trim_spaces":
            text = ' '.join(text.split())
    return text

def extract_concatenate(elements):
    texts = [element.get_text(strip=True) for element in elements]
    return "\n".join(texts)

def extract_code_blocks(elements):
    return [element.get_text() for element in elements]

def extract_list(elements, attributes):
    if attributes:
        data = []
        for element in elements:
            element_data = []
            for attribute in attributes:
                if isinstance(attribute, dict):
                    attr_name = attribute['name']
                    attr_transformations = attribute.get('transformations', [])
                    if attr_name == "text":
                        value = element.get_text(strip=True)
                    else:
                        value = element.get(attr_name)
                    value = apply_transformations(value, attr_transformations)
                    element_data.append(value)
                else:
                    if attribute == "text":
                        element_data.append(element.get_text(strip=True))
                    else:
                        element_data.append(element.get(attribute))
            data.extend(element_data)
    else:
        data = [element.get_text(strip=True) for element in elements]
    return data

def extract_hash(elements, key_attr, value_attr):
    data = {}
    for element in elements:
        key = element.get_text(strip=True) if key_attr == 'text' else element.get(key_attr)
        value = element.get(value_attr)
        data[key] = value
    return data

def extract_single(elements):
    return elements[0].get_text(strip=True) if elements else None

def extract_table(elements):
    headers = [th.get_text(strip=True) for th in elements[0].find_all('th')] if elements else []
    rows = []
    for element in elements:
        row = [td.get_text(strip=True) for td in element.find_all('td')]
        rows.append(row)
    return {"headers": headers, "rows": rows}

def extract_data(soup, config, debug=False):
    def process_selector(soup, selector_config, depth=0):
        title = selector_config.get('title')
        selector = selector_config.get('selector')
        
        if debug:
            indent = "  " * depth
            print(f"{indent}Debug: Processing selector '{title}' with selector '{selector}'", file=sys.stderr)

        if not selector:
            print(f"{indent}Warning: Selector for '{title}' is missing or empty. Skipping.", file=sys.stderr)
            return None

        assemble = selector_config.get('assemble', 'list')
        children = selector_config.get('children', [])
        attributes = selector_config.get('attributes', [])
        transformations = selector_config.get('transformations', [])

        elements = soup.select(selector)
        
        if debug:
            print(f"{indent}Debug: Found {len(elements)} elements for selector '{selector}'", file=sys.stderr)

        if assemble == "concatenate":
            data = extract_concatenate(elements)
        elif assemble == "code_blocks":
            data = extract_code_blocks(elements)
        elif assemble == "list":
            data = extract_list(elements, attributes)
        elif assemble == "hash":
            key_attr = selector_config.get('key_attribute', 'text')
            value_attr = selector_config.get('value_attribute', 'href')
            data = extract_hash(elements, key_attr, value_attr)
        elif assemble == "single":
            data = extract_single(elements)
        elif assemble == "table":
            data = extract_table(elements)
        else:
            print(f"{indent}Warning: Unknown assembly method '{assemble}' for '{title}'. Using 'list' as default.", file=sys.stderr)
            data = extract_list(elements, attributes)

        if debug:
            if isinstance(data, (str, list)):
                preview = str(data)[:100]
            elif isinstance(data, dict):
                preview = str(list(data.items())[:3])
            else:
                preview = str(data)
            print(f"{indent}Debug: Assembled data for '{title}': {preview}{'...' if len(preview) > 100 else ''}", file=sys.stderr)

        # Apply transformations
        if transformations:
            if isinstance(data, list):
                data = [apply_transformations(item, transformations) for item in data]
            elif isinstance(data, str):
                data = apply_transformations(data, transformations)

            if debug:
                print(f"{indent}Debug: Applied transformations for '{title}': {transformations}", file=sys.stderr)

        # Process children recursively
        if children:
            if debug:
                print(f"{indent}Debug: Processing {len(children)} child selectors for '{title}'", file=sys.stderr)
            child_data = {}
            for i, child in enumerate(children):
                if debug:
                    print(f"{indent}  Debug: Processing child {i+1}/{len(children)} for '{title}': {child}", file=sys.stderr)
                child_result = process_selector(soup, child, depth + 1)
                if child_result is not None:
                    child_data[child.get('title')] = child_result
            data = child_data
            if debug:
                print(f"{indent}Debug: Finished processing children for '{title}'", file=sys.stderr)

        return data

    extracted = {}
    for selector in config.get('selectors', []):
        result = process_selector(soup, selector)
        if result is not None:
            extracted[selector.get('title')] = result
    return extracted

def str_presenter(dumper, data):
    if '\n' in data:  # check for multiline string
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)

yaml.add_representer(str, str_presenter)

@click.command()
@click.option('--config', '-c', type=click.File('r'), required=True, help='YAML configuration file with selectors.')
@click.option('--input', '-i', type=click.File('r'), help='Input HTML file. If not provided, reads from stdin.')
@click.option('--output', '-o', type=click.File('w'), help='Output YAML file. If not provided, writes to stdout.')
@click.option('--debug', '-d', is_flag=True, help='Enable debug mode for verbose output.')
def extract_content(config, input, output, debug):
    """
    Extract content from HTML based on YAML-configured CSS selectors.
    """
    if debug:
        print("Debug: Starting extraction process", file=sys.stderr)

    # Load YAML configuration
    config_data = yaml.safe_load(config)

    if debug:
        print(f"Debug: Loaded configuration: {config_data}", file=sys.stderr)

    # Read HTML input from file or stdin
    html_input = input.read() if input else sys.stdin.read()

    if debug:
        print(f"Debug: Read {len(html_input)} characters of HTML input", file=sys.stderr)

    # Parse the HTML using BeautifulSoup
    soup = BeautifulSoup(html_input, 'html.parser')

    if debug:
        print("Debug: HTML parsed successfully", file=sys.stderr)

    # Extract data based on configuration
    extracted_data = extract_data(soup, config_data, debug)

    if debug:
        print(f"Debug: Extracted data: {extracted_data}", file=sys.stderr)

    # Output the extracted data as YAML to file or stdout
    yaml_output = yaml.dump(extracted_data, default_flow_style=False, allow_unicode=True)
    if output:
        output.write(yaml_output)
        if debug:
            print(f"Debug: Written output to file: {output.name}", file=sys.stderr)
    else:
        sys.stdout.write(yaml_output)
        if debug:
            print("Debug: Written output to stdout", file=sys.stderr)

if __name__ == '__main__':
    extract_content()