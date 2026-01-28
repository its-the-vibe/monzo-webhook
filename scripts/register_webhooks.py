#!/usr/bin/env python3
"""
Register a webhook URL with all accounts of the currently logged-in Monzo user.

This script will:
1. List all accounts for the logged-in user
2. Register the provided webhook URL with each account

Usage:
    python register_webhooks.py <webhook_url>

Example:
    python register_webhooks.py https://example.com/webhook

Environment Variables:
    MONZO_ACCESS_TOKEN - Monzo API access token (required)
"""

import os
import sys
from urllib.parse import urlparse
import requests


MONZO_API_BASE = "https://api.monzo.com"
REQUEST_TIMEOUT = 30  # seconds


def get_access_token():
    """Get Monzo access token from environment variable."""
    token = os.environ.get("MONZO_ACCESS_TOKEN")
    if not token:
        print("Error: MONZO_ACCESS_TOKEN environment variable is not set", file=sys.stderr)
        print("\nPlease set your Monzo access token:", file=sys.stderr)
        print("  export MONZO_ACCESS_TOKEN=your_token_here", file=sys.stderr)
        sys.exit(1)
    return token


def list_accounts(access_token):
    """List all accounts for the logged-in user."""
    url = f"{MONZO_API_BASE}/accounts"
    headers = {"Authorization": f"Bearer {access_token}"}
    
    try:
        response = requests.get(url, headers=headers, timeout=REQUEST_TIMEOUT)
        response.raise_for_status()
        data = response.json()
        return data.get("accounts", [])
    except requests.exceptions.RequestException as e:
        print(f"Error listing accounts: {e}", file=sys.stderr)
        sys.exit(1)


def register_webhook(access_token, account_id, webhook_url):
    """Register a webhook for a specific account."""
    url = f"{MONZO_API_BASE}/webhooks"
    headers = {"Authorization": f"Bearer {access_token}"}
    data = {
        "account_id": account_id,
        "url": webhook_url
    }
    
    try:
        response = requests.post(url, headers=headers, data=data, timeout=REQUEST_TIMEOUT)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error registering webhook for account {account_id}: {e}", file=sys.stderr)
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}", file=sys.stderr)
        return None


def main():
    """Main function to register webhooks."""
    # Check command-line arguments
    if len(sys.argv) != 2:
        print("Usage: python register_webhooks.py <webhook_url>", file=sys.stderr)
        print("\nExample:", file=sys.stderr)
        print("  python register_webhooks.py https://example.com/webhook", file=sys.stderr)
        sys.exit(1)
    
    webhook_url = sys.argv[1]
    
    # Validate URL format
    if not webhook_url.startswith(("http://", "https://")):
        print("Error: Webhook URL must start with http:// or https://", file=sys.stderr)
        sys.exit(1)
    
    # Validate URL structure
    try:
        parsed = urlparse(webhook_url)
        if not parsed.netloc:
            print("Error: Invalid webhook URL format", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"Error: Invalid webhook URL: {e}", file=sys.stderr)
        sys.exit(1)
    
    print("Monzo Webhook Registration Tool")
    print("=" * 50)
    print(f"Webhook URL: {webhook_url}")
    
    # Get access token
    access_token = get_access_token()
    
    # List all accounts
    print("\nFetching accounts...")
    accounts = list_accounts(access_token)
    
    if not accounts:
        print("No accounts found.")
        return
    
    print(f"Found {len(accounts)} account(s)")
    
    # Track statistics
    registered_count = 0
    failed_count = 0
    
    # Process each account
    for account in accounts:
        account_id = account.get("id")
        account_desc = account.get("description", "Unknown")
        
        print(f"\nProcessing account: {account_desc} ({account_id})")
        
        result = register_webhook(access_token, account_id, webhook_url)
        
        if result:
            webhook_id = result.get("webhook", {}).get("id", "Unknown")
            print(f"  ✓ Successfully registered webhook")
            print(f"    Webhook ID: {webhook_id}")
            registered_count += 1
        else:
            print(f"  ✗ Failed to register webhook")
            failed_count += 1
    
    # Print summary
    print("\n" + "=" * 50)
    print(f"Summary:")
    print(f"  Successfully registered: {registered_count}")
    print(f"  Failed to register: {failed_count}")
    print("=" * 50)


if __name__ == "__main__":
    main()
