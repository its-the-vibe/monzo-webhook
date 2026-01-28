# Helper Scripts for Monzo Webhook Management

This directory contains Python helper scripts to manage Monzo webhooks across all your accounts.

## Prerequisites

- Python 3.9 or higher
- A valid Monzo API access token
- Install required dependencies:
  ```bash
  pip install -r requirements.txt
  ```

## Scripts

### 1. Delete Webhooks (`delete_webhooks.py`)

Deletes all existing webhooks from all accounts of the currently logged-in Monzo user.

**Usage:**
```bash
export MONZO_ACCESS_TOKEN=your_token_here
python scripts/delete_webhooks.py
```

**What it does:**
1. Lists all accounts for the logged-in user
2. For each account, lists all registered webhooks
3. Deletes each webhook found
4. Displays a summary of deletions

**Example output:**
```
Monzo Webhook Deletion Tool
==================================================

Fetching accounts...
Found 2 account(s)

Processing account: Personal Account (acc_00009abc...)
  Found 3 webhook(s)
    Deleting webhook: webhook_00009xyz...
      URL: https://example.com/webhook
      ✓ Successfully deleted
    ...

==================================================
Summary:
  Total webhooks found: 5
  Successfully deleted: 5
  Failed to delete: 0
==================================================
```

### 2. Register Webhooks (`register_webhooks.py`)

Registers a webhook URL with all accounts of the currently logged-in Monzo user.

**Usage:**
```bash
export MONZO_ACCESS_TOKEN=your_token_here
python scripts/register_webhooks.py <webhook_url>
```

**Example:**
```bash
python scripts/register_webhooks.py https://example.com/webhook
```

**With basic authentication (use with caution):**
```bash
python scripts/register_webhooks.py https://username:password@example.com/webhook
```

**Note:** Embedding credentials in URLs can expose them in logs and browser history. Use this approach only in secure, controlled environments. For production, ensure the webhook endpoint uses HTTPS and consider using alternative authentication methods provided by your webhook server.

**What it does:**
1. Lists all accounts for the logged-in user
2. Registers the provided webhook URL with each account
3. Displays a summary of registrations

**Example output:**
```
Monzo Webhook Registration Tool
==================================================
Webhook URL: https://example.com/webhook

Fetching accounts...
Found 2 account(s)

Processing account: Personal Account (acc_00009abc...)
  ✓ Successfully registered webhook
    Webhook ID: webhook_00009xyz...

Processing account: Joint Account (acc_00009def...)
  ✓ Successfully registered webhook
    Webhook ID: webhook_00009uvw...

==================================================
Summary:
  Successfully registered: 2
  Failed to register: 0
==================================================
```

## Getting a Monzo Access Token

To use these scripts, you'll need a Monzo API access token:

1. Log in to the [Monzo Developer Portal](https://developers.monzo.com/)
2. Go to the "Clients" section
3. Create a new OAuth client or use an existing one
4. Follow the OAuth flow to get an access token
5. Export the token as an environment variable:
   ```bash
   export MONZO_ACCESS_TOKEN=your_token_here
   ```

**Note:** Access tokens expire. If you get authentication errors, you may need to refresh your token.

## Common Workflows

### Replace all webhooks with a new URL:
```bash
# First, delete all existing webhooks
python scripts/delete_webhooks.py

# Then, register the new webhook URL
python scripts/register_webhooks.py https://newurl.com/webhook
```

### Quick setup for a new environment:
```bash
# Install dependencies
pip install -r scripts/requirements.txt

# Set your token
export MONZO_ACCESS_TOKEN=your_token_here

# Register webhooks
python scripts/register_webhooks.py https://myserver.com/webhook
```

## Security Notes

- Always use HTTPS URLs for webhooks in production
- Consider using basic authentication in the webhook URL (e.g., `https://user:pass@example.com/webhook`)
- Keep your access token secure and never commit it to version control
- Access tokens should be treated as sensitive credentials

## Error Handling

Both scripts include error handling for common issues:
- Missing or invalid access token
- Network errors
- Invalid API responses
- Invalid webhook URLs

If you encounter errors, check:
1. Your access token is valid and not expired
2. You have an active internet connection
3. The Monzo API is accessible
4. For registration: The webhook URL is valid and uses HTTP/HTTPS
