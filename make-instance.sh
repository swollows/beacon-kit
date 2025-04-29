INSTANCE=beaconkit-dev
ZONE=us-central1-a
MACHINE="c4a-standard-8"

gcloud compute instances create "$INSTANCE" \
  --zone "$ZONE" \
  --machine-type "$MACHINE" \
  --image-family ubuntu-minimal-2404-lts-arm64 \
  --image-project ubuntu-os-cloud \
  --boot-disk-size 200GB
  --boot-sidk-type pd-balanced \
  --tags ssh \
  --metadata=user-data='''#cloud-config
packages:
  - make
  - git
  - vim
  - zip
  - net-tools
  - dnsutils
  - ca-certificates
  - curl
  - docker.io
  - kurtosis-cli
  - go
  - openvpn
  - easy-rsa

runcmd:
  - curl -L https://foundry.paradigm.xyz | bash
  - source /home/jonathan/.bashrc && foundryup
'''