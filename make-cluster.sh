#!/usr/bin/env bash
set -euo pipefail

# ====== 변수만 바꿔서 사용 ======
CLUSTER="beaconkit-dev"
REGION="us-central1"          # us-central1 지역의 a 존
MACHINE="e2-standard-2"       # 필요시 더 작게(e2-medium) 변경

# 1) 클러스터 만들기 ─ 노드 1개
gcloud container clusters create "$CLUSTER" \
  --region "$REGION" \
  --machine-type "$MACHINE" \
  --num-nodes 1 \
  --enable-ip-alias \
  --enable-autorepair \
  --enable-autoupgrade

# 2) 바로 0개로 리사이즈(Cluster Autoscaler X, 단순 resize)
gcloud container clusters resize "$CLUSTER" \
  --region "$REGION" \
  --node-pool default-pool \
  --num-nodes 1 --quiet          # Kurtosis 쓰기 전

# 3) Kubeconfig 컨텍스트 가져오기
gcloud container clusters get-credentials "$CLUSTER" --region "$REGION"

echo "✅  Cluster $CLUSTER ready (node count now 0)."
