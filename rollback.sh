#!/bin/bash

# rollback.sh
# Remove deployed files on the remote server

# SSH key path, user, host, and remote path should match config.json
SSH_KEY="/home/khaledibra/.ssh/id_rsa"
REMOTE_USER="appuser"
REMOTE_HOST="server.example.com"
REMOTE_PATH="/var/www/app"

# Execute the removal command on the remote server
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no "$REMOTE_USER@$REMOTE_HOST" "rm -rf $REMOTE_PATH/*"

# Check if the command was successful
if [ $? -eq 0 ]; then
    echo "Successfully removed deployed files from $REMOTE_PATH on $REMOTE_HOST"
else
    echo "Failed to remove deployed files from $REMOTE_PATH on $REMOTE_HOST"
    exit 1
fi