#!/bin/bash

set -x

echo "Deleting cluster"
curl -H "Apikey:${API_KEY}" -H "ProjectUID:${PROJECT_ID}" --request DELETE  https://${SC_HOST}/v1/spectroclusters/${CLUSTER_ID}
sleep 5
STATE=$(curl -H "Apikey:${API_KEY}" -H "ProjectUID:${PROJECT_ID}" https://${SC_HOST}/v1/spectroclusters/${CLUSTER_ID} | jq '.status.state')
DELETED_STATE="Deleted"

COUNT=0
while [ $COUNT -lt 60 ]
do
if [ "$STATE" == "$DELETED_STATE" ]; then
    echo "Cluster is in $STATE state"
    exit 0
else
    echo "Cluster is in $STATE state"
fi
COUNT=$[$COUNT+1]
sleep 5
done

# call force delete api after 5 mins if cluster isn't deleted
curl --location --request PATCH 'https://${SC_HOST}/v1/spectroclusters/${CLUSTER_ID}/status/conditions' \
--header 'Apikey:${API_KEY}' \
--header 'ProjectUID:${PROJECT_ID}' \
--header 'Content-Type: application/json' \
--data-raw '[
    {
        "message": "cleaned up",
        "reason": "CloudInfrastructureCleanedUp",
        "status": "True",
        "type": "CloudInfrastructureCleanedUp"
    }
]'

sleep 5

curl -H "Apikey:${API_KEY}" -H "ProjectUID:${PROJECT_ID}" --request DELETE  https://${SC_HOST}/v1/spectroclusters/${CLUSTER_ID}