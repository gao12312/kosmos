#!/usr/bin/env bash

if [ -z "$KUBECONFIG" ]; then
  echo "Error: KUBECONFIG_CONTENT is not set."
  exit 1
fi

echo "$KUBECONFIG" > /app/kubeconfig
if [ $? -ne 0 ]; then
  echo "Error: Failed to write kubeconfig to $KUBECONFIG."
  exit 1
fi

echo "KUBECONFIG has been written to $KUBECONFIG."


WEB_USER="$WEB_USER" sed -i 's/^WEB_USER=.*/WEB_USER="'"$WEB_USER"'"/' /app/agent.env
WEB_PASS="$WEB_PASS" sed -i 's/^WEB_PASS=.*/WEB_PASS="'"$WEB_PASS"'"/' /app/agent.env
WEB_PORT="$WEB_PORT" sed -i 's/^WEB_PORT=.*/WEB_PORT="'"$WEB_PORT"'"/' /app/agent.env
NODE_NAME="$NODE_NAME" sed -i 's/^NODE_NAME=.*/NODE_NAME="'"$NODE_NAME"'"/' /app/agent.env

sha256sum /app/node-agent > /app/node-agent.sum
sha256sum /host-path/node-agent >> /app/node-agent.sum
rsync -avz /app/ /host-path/
cp /app/node-agent.service /host-systemd/node-agent.service 

