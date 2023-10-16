# Designing a DSL to configure bleve

## Introduction

This file is mostly a bunch of ChatGPT transcripts as I am designing a DSL to configure bleve. 

```yaml
test: 123
```

- list
- listfoobar
- list 3
  - nested list
  - nested list 2

## Designing a text analysis chain DSL (2023-08-23)

### Attempt 1

Designing a YAML DSL for the text analysis chains in bleve involves structuring the configuration in a way that's
intuitive, extensible, and expressive. Given the modular components of the analysis in bleve, we can break down the YAML
configuration into distinct sections for character filters, tokenizers, and token filters.

Here's a proposed structure for the DSL:

```yaml
analysis:
  character_filters:
    - name: filter_name
      type: filter_type
      options:
        option1: value1
        # ... other options

  tokenizers:
    - name: tokenizer_name
      type: tokenizer_type
      options:
        option1: value1
        # ... other options

  token_filters:
    - name: filter_name
      type: filter_type
      options:
        option1: value1
        # ... other options

  analyzers:
    - name: analyzer_name
      character_filters: [list_of_char_filter_names]
      tokenizer: tokenizer_name
      token_filters: [list_of_token_filter_names]
```

Now, let's provide 5 examples based on the information given:

1. **Simple Analyzer**:

```yaml
analysis:
  tokenizers:
    - name: letter_tokenizer
      type: Letter

  token_filters:
    - name: lowercase_filter
      type: Lowercase

  analyzers:
    - name: simple_analyzer
      tokenizer: letter_tokenizer
      token_filters:
        - lowercase_filter
```

2. **Standard Analyzer**:

```yaml
analysis:
  tokenizers:
    - name: unicode_tokenizer
      type: Unicode

  token_filters:
    - name: lowercase_filter
      type: Lowercase
    - name: english_stop_filter
      type: Stop Token
      options:
        stop_token_map: English

  analyzers:
    - name: standard_analyzer
      tokenizer: unicode_tokenizer
      token_filters:
        - lowercase_filter
        - english_stop_filter
```

3. **Keyword Analyzer**:

```yaml
analysis:
  analyzers:
    - name: keyword_analyzer
      tokenizer: Single Token
```

4. **Detect Language Analyzer**:

```yaml
analysis:
  tokenizers:
    - name: single_token_tokenizer
      type: Single Token

  token_filters:
    - name: lowercase_filter
      type: Lowercase
    - name: cld2_filter
      type: CLD2

  analyzers:
    - name: detect_language_analyzer
      tokenizer: single_token_tokenizer
      token_filters:
        - lowercase_filter
        - cld2_filter
```

5. **English with HTML Character Filter**:

```yaml
analysis:
  character_filters:
    - name: html_filter
      type: HTML

  tokenizers:
    - name: unicode_tokenizer
      type: Unicode

  token_filters:
    - name: lowercase_filter
      type: Lowercase
    - name: english_stop_filter
      type: Stop Token
      options:
        stop_token_map: English

  analyzers:
    - name: english_html_analyzer
      character_filters:
        - html_filter
      tokenizer: unicode_tokenizer
      token_filters:
        - lowercase_filter
        - english_stop_filter
```

Note that these examples show a subset of the full configuration options. The designed DSL is intended to be extensible
and flexible to handle other configurations not shown here.

### Attempt 2 (regenerate)