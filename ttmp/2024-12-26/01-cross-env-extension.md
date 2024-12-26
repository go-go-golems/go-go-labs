To adapt a Chrome extension for Firefox, you'll need to make several key modifications:

1. Manifest File Changes:
- Firefox uses manifest v2 (Chrome now uses v3)
- Replace "chrome" namespace with "browser" namespace in your code
- Modify the manifest.json to use Firefox-specific properties:
  ```json
  {
    "browser_specific_settings": {
      "gecko": {
        "id": "your-extension@your-domain.com",
        "strict_min_version": "42.0"
      }
    }
  }
  ```

2. API Adjustments:
- Use the `browser` namespace instead of `chrome`
- Firefox requires promises instead of callbacks for most APIs
- Example change:
  ```javascript
  // Chrome
  chrome.tabs.query({active: true}, function(tabs) {
    // callback
  });

  // Firefox
  await browser.tabs.query({active: true});
  ```

3. WebExtension Polyfill:
If you want to maintain a single codebase, you can use the WebExtension browser API Polyfill (webextension-polyfill) which allows you to write Firefox-style promise-based code that works in both browsers.

4. Permissions:
- Review your permissions as Firefox might have different requirements
- Some Chrome-specific APIs might not be available in Firefox

5. Extension Storage:
- Firefox uses slightly different storage mechanisms
- Ensure you're using compatible storage APIs

6. Testing:
- Test in Firefox's about:debugging page
- Use web-ext tool for development and testing
- Verify all features work as expected in Firefox

---

I'll help you set up a build system that can generate both Chrome and Firefox extensions from the same source code. Here's a complete solution:

```javascript
// package.json snippet for development scripts
{
  "scripts": {
    "build:chrome": "cross-env BROWSER=chrome webpack --config webpack.config.js",
    "build:firefox": "cross-env BROWSER=firefox webpack --config webpack.config.js",
    "build": "npm run build:chrome && npm run build:firefox"
  }
}
```

Here's the recommended project structure:

```
src/
  ├── manifest/
  │   ├── chrome.json    # Chrome-specific manifest
  │   └── firefox.json   # Firefox-specific manifest
  ├── common/            # Shared code
  ├── background/
  └── content/
webpack.config.js
package.json
```

Key implementation approaches:

1. Use webpack with environment variables:
```javascript
// webpack.config.js
const browser = process.env.BROWSER || 'chrome';

module.exports = {
  entry: {
    background: './src/background/index.js',
    content: './src/content/index.js'
  },
  output: {
    path: path.resolve(__dirname, `dist/${browser}`),
    filename: '[name].js'
  },
  plugins: [
    new CopyWebpackPlugin({
      patterns: [
        {
          from: `src/manifest/${browser}.json`,
          to: 'manifest.json',
          transform(content) {
            return generateManifest(content, browser);
          }
        }
      ]
    })
  ]
};
```

2. Use the webextension-polyfill:
```javascript
// src/common/browser-polyfill.js
import browser from 'webextension-polyfill';

export const getBrowser = () => {
  return typeof chrome !== 'undefined' ? chrome : browser;
};
```

3. Create abstraction layers for browser-specific code:
```javascript
// src/common/api.js
import { getBrowser } from './browser-polyfill';

export const storage = {
  async get(key) {
    const browser = getBrowser();
    return browser.storage.local.get(key);
  },
  // ... other storage methods
};
```

4. Handle manifest differences:
```javascript
// utils/manifest.js
function generateManifest(content, browser) {
  const manifest = JSON.parse(content);
  
  if (browser === 'firefox') {
    manifest.browser_specific_settings = {
      gecko: {
        id: "your-extension@domain.com"
      }
    };
  }
  
  // Handle manifest version differences
  manifest.manifest_version = browser === 'chrome' ? 3 : 2;
  
  return JSON.stringify(manifest, null, 2);
}
```

5. Set up conditional imports:
```javascript
// webpack.config.js additional rules
module: {
  rules: [
    {
      test: /\.js$/,
      loader: 'string-replace-loader',
      options: {
        search: '__BROWSER__',
        replace: browser,
      }
    }
  ]
}
```

Development workflow:

1. Install dependencies:
```bash
npm install --save-dev webpack webpack-cli cross-env webextension-polyfill
```

2. Run builds:
```bash
# Build for both browsers
npm run build

# Build for specific browser
npm run build:chrome
npm run build:firefox
```

This setup will:
- Generate separate builds for Chrome and Firefox in `dist/chrome` and `dist/firefox`
- Handle manifest differences automatically
- Use the same source code with browser-specific abstractions
- Make maintenance easier with shared code
- Support different API patterns between browsers
