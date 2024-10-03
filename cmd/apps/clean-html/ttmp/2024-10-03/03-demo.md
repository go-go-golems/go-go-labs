```
❯ llmwebc --context test/sample/sample.html --context extract-dsl.md "Create a YAML to extract the important information out of this. First list the differtent types of information, then filter out the important one, then the YAML."
```

> **Types of Information:**
> 1. Title
> 2. Description
> 3. Links (API documentation, Source Code, Contributing, Code of Conduct)
> 4. Code Snippets
> 5. Policies
> 6. Statistics (Stars, Network)
> 7. Contributors
> 8. External Links (Docs, Blog, Partners)
> 
> **Important Information:**
> - Title
> - Description
> - Links
> - Code Snippets
> - Statistics
> - Contributors
> 
> **YAML:**
> 
> ```yaml
> selectors:
  > - title: Page Title
    > selector: h1.heading-element
    > assemble: single
  > - title: Description
    > selector: .markdown-body p
    > assemble: single
  > - title: Links
    > selector: .markdown-body ul li a
    > assemble: hash
    > key_attribute: text
    > value_attribute: href
  > - title: Code Snippets
    > selector: .highlight pre
    > assemble: code_blocks
  > - title: Statistics
    > selector: .statistics
    > assemble: hash
    > children:
      > - title: Stars
        > selector: #repo-stars-counter
        > assemble: single
      > - title: Network
        > selector: #repo-network-counter
        > assemble: single
  > - title: Contributors
    > selector: .contributors img
    > assemble: list
    > attributes:
      > - alt

```
❯ cat test/sample/sample.html
```

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sample Repository Page</title>
</head>
<body>
    <div class="repository-content">
        <h1 class="heading-element">Pundit</h1>
        <div class="markdown-body">
            <p>Pundit provides a set of helpers which guide you in leveraging regular Ruby classes and object oriented design patterns to build a straightforward, robust, and scalable authorization system.</p>
            <ul>
                <li><a href="https://example.com/api">API documentation for the most recent version</a></li>
                <li><a href="https://github.com/example/repo">Source Code</a></li>
                <li><a href="https://github.com/example/repo/contributing">Contributing</a></li>
                <li><a href="https://github.com/example/repo/code_of_conduct">Code of Conduct</a></li>
            </ul>
            <div class="highlight">
                <pre>def example_function():
    print("Hello, world!")</pre>
            </div>
            <div class="highlight">
                <pre>class ApplicationController < ActionController::Base
  include Pundit::Authorization
end</pre>
            </div>
            <h2 class="heading-element">Policies</h2>
            <ul>
                <li>Actions</li>
                <li>Security</li>
                <li>Codespaces</li>
                <li>GitHub Copilot</li>
                <li>Code review</li>
                <li>Issues</li>
                <li>Discussions</li>
            </ul>
        </div>
        <div class="statistics">
            <span id="repo-stars-counter">8.3k</span>
            <span id="repo-network-counter">629</span>
            <div class="contributors">
                <img class="avatar" alt="Contributor1" src="contributor1.jpg">
                <img class="avatar" alt="Contributor2" src="contributor2.jpg">
                <img class="avatar" alt="Contributor3" src="contributor3.jpg">
            </div>
        </div>
    </div>
    <div class="HeaderMenu-dropdown">
        <a class="HeaderMenu-dropdown-link Link--external" href="https://github.com/docs">Docs</a>
        <a class="HeaderMenu-dropdown-link Link--external" href="https://github.com/blog">Blog</a>
        <a class="HeaderMenu-dropdown-link Link--external" href="https://github.com/partners">Partners</a>
    </div>
</body>
</html>%                                                                                                                                                           
```

```
❯ python extract_html.py --config test/sample/info.yaml -i test/sample/sample.html   
```

```yaml
Code Snippets:
- "|\ndef example_function():\n    print(\"Hello, world!\")"
- "|\nclass ApplicationController < ActionController::Base\n  include Pundit::Authorization\n\
  end"
Contributors:
- Contributor1
- Contributor2
- Contributor3
Description: Pundit provides a set of helpers which guide you in leveraging regular
  Ruby classes and object oriented design patterns to build a straightforward, robust,
  and scalable authorization system.
Links:
  API documentation for the most recent version: https://example.com/api
  Code of Conduct: https://github.com/example/repo/code_of_conduct
  Contributing: https://github.com/example/repo/contributing
  Source Code: https://github.com/example/repo
Page Title: Pundit
Statistics:
  Network: '629'
  Stars: 8.3k
```