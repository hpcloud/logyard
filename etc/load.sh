#!/bin/sh
# simple script to load config into doozer until 'kato' does it
# itself.

set -xe

ruby -ryaml -rjson -e 'puts YAML.load_file("etc/logyard.yml").to_json'  | redis-cli -p 5454 -x set config:logyard
ruby -ryaml -rjson -e 'puts YAML.load_file("etc/apptail.yml").to_json'  | redis-cli -p 5454 -x set config:apptail
ruby -ryaml -rjson -e 'puts YAML.load_file("etc/systail.yml").to_json'  | redis-cli -p 5454 -x set config:systail
ruby -ryaml -rjson -e 'puts YAML.load_file("etc/cloud_events.yml").to_json'  | redis-cli -p 5454 -x set config:cloud_events
