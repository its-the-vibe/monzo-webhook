#!/usr/bin/env python3
"""
Delete all webhooks for all accounts of the currently logged-in Monzo user.

This script will:
1. List all accounts for the logged-in user
2. For each account, list all registered webhooks
3. Delete each webhook found

Usage:
    python delete_webhooks.py

Environment Variables:
    MONZO_ACCESS_TOKEN - Monzo API access token (required)
"""

import os
import sys
import requests


MONZO_API_BASE = "https://api.monzo.com"


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
        response = requests.get(url, headers=headers)
        response.raise_for_status()
        data = response.json()
        return data.get("accounts", [])
    except requests.exceptions.RequestException as e:
        print(f"Error listing accounts: {e}", file=sys.stderr)
        sys.exit(1)


def list_webhooks(access_token, account_id):
    """List all webhooks for a specific account."""
    url = f"{MONZO_API_BASE}/webhooks"
    headers = {"Authorization": f"Bearer {access_token}"}
    params = {"account_id": account_id}
    
    try:
        response = requests.get(url, headers=headers, params=params)
        response.raise_for_status()
        data = response.json()
        return data.get("webhooks", [])
    except requests.exceptions.RequestException as e:
        print(f"Error listing webhooks for account {account_id}: {e}", file=sys.stderr)
        return []


def delete_webhook(access_token, webhook_id):
    """Delete a specific webhook."""
    url = f"{MONZO_API_BASE}/webhooks/{webhook_id}"
    headers = {"Authorization": f"Bearer {access_token}"}
    
    try:
        response = requests.delete(url, headers=headers)
        response.raise_for_status()
        return True
    except requests.exceptions.RequestException as e:
        print(f"Error deleting webhook {webhook_id}: {e}", file=sys.stderr)
        return False


def main():
    """Main function to delete all webhooks."""
    print("Monzo Webhook Deletion Tool")
    print("=" * 50)
    
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
    total_webhooks = 0
    deleted_webhooks = 0
    
    # Process each account
    for account in accounts:
        account_id = account.get("id")
        account_desc = account.get("description", "Unknown")
        
        print(f"\nProcessing account: {account_desc} ({account_id})")
        
        # List webhooks for this account
        webhooks = list_webhooks(access_token, account_id)
        
        if not webhooks:
            print(f"  No webhooks found for this account")
            continue
        
        print(f"  Found {len(webhooks)} webhook(s)")
        total_webhooks += len(webhooks)
        
        # Delete each webhook
        for webhook in webhooks:
            webhook_id = webhook.get("id")
            webhook_url = webhook.get("url", "Unknown URL")
            
            print(f"    Deleting webhook: {webhook_id}")
            print(f"      URL: {webhook_url}")
            
            if delete_webhook(access_token, webhook_id):
                print(f"      ✓ Successfully deleted")
                deleted_webhooks += 1
            else:
                print(f"      ✗ Failed to delete")
    
    # Print summary
    print("\n" + "=" * 50)
    print(f"Summary:")
    print(f"  Total webhooks found: {total_webhooks}")
    print(f"  Successfully deleted: {deleted_webhooks}")
    print(f"  Failed to delete: {total_webhooks - deleted_webhooks}")
    print("=" * 50)


if __name__ == "__main__":
    main()
