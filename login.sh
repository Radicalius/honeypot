#!/bin/bash

. secrets.env

ssh -i "$PEM_FILE" "$USERNAME@$HOST_IP"