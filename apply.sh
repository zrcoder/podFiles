#!/bin/bash

# Exit on error and undefined variables
set -eu

# Check terminal color support
if [ -t 1 ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors)" -ge 8 ]; then
    # Define colors for terminal output
    GREEN='\033[0;32m'  # Success
    BLUE='\033[0;34m'   # Info
    YELLOW='\033[1;33m' # Warning
    RED='\033[0;31m'    # Error
    NC='\033[0m'        # Reset
else
    # No color support
    GREEN='' BLUE='' YELLOW='' RED='' NC=''
fi

# Create temporary files
create_yaml_files() {
    cat > cluster-resources.yaml << 'EOF'
# RBAC Role
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: podfiles-role
rules:
  - apiGroups: [""]
    resources:
      - "namespaces"
      - "pods"
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create", "get"]
---
# RBAC RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: podfiles-role-binding
subjects:
  - kind: ServiceAccount
    name: podfiles-sa
    namespace: ${NAMESPACE}
roleRef:
  kind: ClusterRole
  name: podfiles-role
  apiGroup: rbac.authorization.k8s.io

EOF

    cat > namespace-resources.yaml << 'EOF'
# ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: podfiles-sa

---
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: podfiles
  labels:
    app: podfiles
spec:
  replicas: 1
  selector:
    matchLabels:
      app: podfiles
  template:
    metadata:
      labels:
        app: podfiles
    spec:
      serviceAccountName: podfiles-sa
      containers:
        - name: podfiles
          image: ${IMAGE}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              name: http
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10

---
# Service
apiVersion: v1
kind: Service
metadata:
  name: podfiles
spec:
  selector:
    app: podfiles
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ${SERVICE_TYPE}

EOF

    cat > ingress.yaml << 'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: podfiles
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
    - host: ${INGRESS_DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: podfiles
                port:
                  number: 80

EOF
}

# Cleanup temporary files
cleanup() {
    rm -f cluster-resources.yaml namespace-resources.yaml ingress.yaml
}

trap cleanup EXIT
trap 'if [ -n "$RED" ]; then echo -e "${RED}[ERROR]${NC} Line $LINENO, exit code: $?"; else echo "[ERROR] Line $LINENO, exit code: $?"; fi' ERR

# Create yaml files
create_yaml_files

# Rest of your existing script...
