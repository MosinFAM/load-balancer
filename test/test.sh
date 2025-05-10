#!/bin/bash

echo "Running tests..."

for i in {1..6}
do
  curl -s http://localhost:8080/ -w "\n"
done
