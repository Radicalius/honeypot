#!/bin/bash

. secrets.env

ssh -i "$PEM_FILE" -p $PORT "$USERNAME@$HOST_IP"