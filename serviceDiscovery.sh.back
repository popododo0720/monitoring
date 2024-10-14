#!/bin/bash
OPENRC_FILE="/etc/kolla/admin-openrc"

if [ -f "$OPENRC_FILE" ]; then
  source "$OPENRC_FILE"
else
  echo "Error: $OPENRC_FILE not found."
  exit 1
fi

# get token id
TOKEN_ID=$(openstack token issue | grep '^| id' | awk -F '|' '{print $3}' | xargs)

# get instance json and format
curl -s -H "X-Auth-Token: $TOKEN_ID" https://192.168.0.60:8774/v2.1/servers/detail | \
        jq -r '
        [
                {
                        targets: ([
                                .servers[] |
                                select(.addresses["internal-net"] != null) |
                                .addresses["internal-net"][] |
                                .addr + ":8088"
                        ] +
			[
            			.servers[] |
            			select(.addresses["external-net"] != null) |
            			.addresses["external-net"][] |
            			.addr + ":8088"
			])
                }
        ]' > /monitoring/targets.json

cat /monitoring/targets.json

docker cp /monitoring/targets.json prometheus_server:/

