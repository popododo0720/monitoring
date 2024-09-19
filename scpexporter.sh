#!/bin/bash

KEY_PATH="/home/coremax/test.pem"
LOCAL_FILE="/home/coremax/custom_exporter"
REMOTE_USER="ubuntu"
REMOTE_HOST="192.168.0.249"
REMOTE_PATH="/home/ubuntu/"

scp -i "$KEY_PATH" "$LOCAL_FILE" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH"

if [ $? -ne 0 ]; then
    echo "File transfer failed."
    exit 1
fi

ssh -i "$KEY_PATH" "$REMOTE_USER@$REMOTE_HOST" << 'EOF'
    cd /home/ubuntu
    if [ -f "custom_exporter" ]; then
        ./custom_exporter &
    else
        echo "custom_exporter file not found."
        exit 1
    fi
EOF

if [ $? -ne 0 ]; then
    echo "SSH command failed."
    exit 1
fi


