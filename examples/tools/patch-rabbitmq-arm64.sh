#!/usr/bin/env sh
set -euxo pipefail
# 1. set proper config file path using the bitnamit rabbitmq config
echo "1. configure statefulset"

cat <<\EOF | kubectl patch statefulsets orchestrator-org-1-rabbitmq -n org-1 --patch "$(cat -)"
---
spec:
  template:
    spec:
      containers:
      - name: rabbitmq
        env:
          - name: RABBITMQ_CONFIG_FILE
            value: "/bitnami/rabbitmq/conf/rabbitmq.conf"
EOF

# apply changes
kubectl delete pod orchestrator-org-1-rabbitmq-0 -n org-1
# wait for the rabbitmq pod to be back up
kubectl wait $(kubectl get -n org-1 pod -l app.kubernetes.io/name=rabbitmq,app.kubernetes.io/instance=orchestrator-org-1 -o name) -n org-1 --for=condition=ready

# 2. authenticate user

echo "2. changing user password"

kubectl exec $(kubectl get -n org-1 pod -l app.kubernetes.io/name=rabbitmq,app.kubernetes.io/instance=orchestrator-org-1 -o name) -n org-1 rabbitmqctl change_password user password

# 3. enable rabbitmq management plugin for the rabbitmq operator to register users, queues, exchanges and channels
echo "3. enable rabbitmq management plugin"

kubectl exec $(kubectl get -n org-1 pod -l app.kubernetes.io/name=rabbitmq,app.kubernetes.io/instance=orchestrator-org-1 -o name) -n org-1 -- rabbitmq-plugins enable rabbitmq_management
