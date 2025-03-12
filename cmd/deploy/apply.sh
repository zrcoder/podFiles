#!/bin/bash

# Exit on error, undefined variables
set -eu

# Error handler
trap 'echo "❌ Error on line $LINENO. Exit code: $?"' ERR

# Default values
DEFAULT_NAMESPACE="podfiles"
DEFAULT_IMAGE="podfiles:latest"
DEFAULT_SERVICE_TYPE="ClusterIP"

# Helper function to read user input with default value
read_input() {
    local prompt="$1"
    local default="$2"
    local input
    printf "%s [%s]: " "$prompt" "$default" >/dev/tty
    read input </dev/tty
    echo "${input:-$default}"
}

# Helper function to read service type
read_service_type() {
    local service_type
    printf "Choose service type [1]: (1) ClusterIP | (2) NodePort | (3) Ingress\n" >/dev/tty
    printf "> " >/dev/tty
    read choice </dev/tty
    choice="${choice:-1}"
    case $choice in
        1|"") echo "ClusterIP";;
        2) echo "NodePort";;
        3) echo "Ingress";;
        *) 
            printf "Invalid choice, using default: ClusterIP\n" >/dev/tty
            echo "ClusterIP";;
    esac
}

# Interactive inputs
echo "Welcome to podFiles deployment script!"
echo "Press enter to accept default values or input your custom values."
echo

NAMESPACE=$(read_input "Enter namespace" "$DEFAULT_NAMESPACE")
IMAGE=$(read_input "Enter image" "$DEFAULT_IMAGE")
SERVICE_TYPE=$(read_service_type)

# Handle Ingress domain if needed
INGRESS_DOMAIN=""
if [ "$SERVICE_TYPE" = "Ingress" ]; then
    INGRESS_DOMAIN=$(read_input "Enter ingress domain" "podfiles.example.com")
    SERVICE_TYPE="ClusterIP"  # 实际 service 类型为 ClusterIP
fi

# Display configuration for confirmation
echo
echo "Configuration Summary:"
echo "  Namespace: $NAMESPACE"
echo "  Image: $IMAGE"
echo "  Service Type: $SERVICE_TYPE"
if [ -n "$INGRESS_DOMAIN" ]; then
    echo "  Ingress Domain: $INGRESS_DOMAIN"
fi

# Ask for confirmation
echo
read -p "Proceed with deployment? [Y/n]: " confirm
confirm="${confirm:-y}"
if [[ ! $confirm =~ ^[Yy] ]]; then
    echo "Deployment cancelled."
    exit 0
fi

# Create namespace if it doesn't exist
kubectl get namespace $NAMESPACE > /dev/null 2>&1 || kubectl create namespace $NAMESPACE

# Apply cluster resources
echo "Applying cluster resources..."
NAMESPACE=$NAMESPACE SERVICE_TYPE=$SERVICE_TYPE envsubst < cluster-resources.yaml | kubectl apply -f -

# Apply namespace resources
echo "Applying namespace resources..."
IMAGE=$IMAGE envsubst < namespace-resources.yaml | kubectl apply -n $NAMESPACE -f -

# Apply ingress if domain is provided
if [ -n "$INGRESS_DOMAIN" ]; then
    echo "Applying ingress..."
    INGRESS_DOMAIN=$INGRESS_DOMAIN envsubst < ingress.yaml | kubectl apply -n $NAMESPACE -f -
fi

echo "✅ Deployment completed successfully"
echo
echo "Access podFiles:"
if [ -n "$INGRESS_DOMAIN" ]; then
    echo "  https://$INGRESS_DOMAIN"
elif [ "$SERVICE_TYPE" = "NodePort" ]; then
    NODE_PORT=$(kubectl get svc podfiles -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
    echo "  Service is exposed on NodePort: $NODE_PORT"
    echo "  Access URLs:"
    kubectl get nodes -o jsonpath='{range .items[*]}{.status.addresses[?(@.type=="InternalIP")].address}{"\n"}{end}' | \
    while read node_ip; do
        echo "  http://$node_ip:$NODE_PORT"
    done
else
    echo "  Use port-forward to access:"
    echo "  kubectl port-forward -n $NAMESPACE svc/podfiles 8080:80"
fi
