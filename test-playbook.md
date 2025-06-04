# Python Coding Standards

## Overview
This document outlines the coding standards for Python development.

## Rules
1. Use snake_case for variable names
2. Use PascalCase for class names
3. Maximum line length is 88 characters
4. Use type hints for all function parameters

## Examples
```python
def calculate_total(items: List[Item]) -> float:
    return sum(item.price for item in items)
```
