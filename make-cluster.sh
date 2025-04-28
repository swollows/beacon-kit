REGION="us-central1"
CLUSTER="beaconkit-dev"

gcloud container clusters create "${CLUSTER}" \
	--region "${REGION}" \
	--num-nodes "2" \
	--machine-type "e2-standard-4" \
	--disk-size "50" \
    --release-channel "regular" \
	--enable-ip-alias \
	--labels=env=dev,app=beaconkit

gcloud container clusters get-credentials "${CLUSTER}" --region "${REGION}"
