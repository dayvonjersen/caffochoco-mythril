#!/bin/bash

ruby -ryaml -rjson -e 'puts JSON.generate(YAML.load(ARGF))' < data.yaml > data.json
