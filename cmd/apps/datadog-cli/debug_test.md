# Debugging Datadog CLI Authentication

The Datadog CLI now has comprehensive logging to help debug authentication issues. Here's how to use it:

## Enable Debug Logging

Run any command with `--log-level debug` to see detailed logging:

```bash
# Test authentication with debug logging
export DATADOG_CLI_API_KEY="your-api-key"
export DATADOG_CLI_APP_KEY="your-app-key"
export DATADOG_CLI_SITE="datadoghq.com"

go run ./go-go-labs/cmd/apps/datadog-cli logs query "status:error" --log-level debug --limit 1
```

## What to Look For

The debug output will show:

### 1. Environment Variable Loading
```
DEBUG environment variables status api_key_set=true app_key_set=true site="datadoghq.com"
```

### 2. Parameter Layer Processing
```
DEBUG Datadog settings extracted successfully api_key_set=true app_key_set=true site="datadoghq.com"
DEBUG Validating Datadog settings api_key_set=true app_key_prefix="ab12****"
```

### 3. Client Creation
```
DEBUG Creating Datadog client api_key_prefix="ab12****" app_key_prefix="cd34****" site="datadoghq.com"
DEBUG Set Datadog API host host="https://api.datadoghq.com"
```

### 4. Authentication Test
```
DEBUG Testing authentication with Datadog API
INFO Datadog API authentication successful status_code=200
```

### 5. Query Execution
```
DEBUG Starting Datadog logs search execution query="status:error" from=... to=... limit=1
DEBUG Executing logs search page page=1 cursor="none"
DEBUG Received logs search response http_status=200 page=1
```

## Common Issues

### Empty API Keys
If you see:
```
ERROR API key is missing - check DATADOG_CLI_API_KEY environment variable
```
Make sure you've set the environment variable correctly.

### Wrong Site
If you're using Datadog EU or other sites, set:
```bash
export DATADOG_CLI_SITE="datadoghq.eu"  # for EU
```

### Authentication Failures
If you see:
```
ERROR Authentication test failed status_code=403
```
Your API key or app key might be invalid or lack permissions.

### Network Issues
If you see connection errors, check:
- Network connectivity to api.datadoghq.com (or your site)
- Firewall settings
- Proxy configuration

## Test Environment Variables

Check if your environment variables are properly set:

```bash
echo "API Key set: $([ -n "$DATADOG_CLI_API_KEY" ] && echo "YES" || echo "NO")"
echo "App Key set: $([ -n "$DATADOG_CLI_APP_KEY" ] && echo "YES" || echo "NO")"
echo "Site: ${DATADOG_CLI_SITE:-datadoghq.com}"
```

## Minimal Test

Try this minimal test to isolate authentication issues:

```bash
# Set your credentials
export DATADOG_CLI_API_KEY="your-actual-api-key"
export DATADOG_CLI_APP_KEY="your-actual-app-key"

# Test with minimal query and debug logging
go run ./go-go-labs/cmd/apps/datadog-cli logs query "*" --limit 1 --log-level debug
```

The debug logs will help identify exactly where the authentication process is failing.
