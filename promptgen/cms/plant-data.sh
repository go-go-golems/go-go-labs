#!/usr/bin/env bash

cd cmd/cms/pkg

echo Plant schema definition:
cat test-data/05-plant.yaml
echo
echo

echo Test plants:

for i in test-data/05-plants/*; do
    echo $i:
    cat $i
done
