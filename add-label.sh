#!/bin/bash

set -euo pipefail

for file in *.yaml
do
  ytt -f ytt.yml -f "$file" > tkg14-"$file"
done

rm -f dashboard-*

for file in *.yaml
do
  mv "$file" "${file:6}"
done
