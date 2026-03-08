#!/bin/bash

echo "Starting Consul..."
consul agent -dev &

echo "Starting Nacos..."
cd ~/dev/nacos
sh bin/startup.sh -m standalone &

echo "All services started."
