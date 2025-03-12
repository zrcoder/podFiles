# PodFiles

PodFiles is a Go - based project designed to interact with Kubernetes clusters. It features a concise UI that simplifies the user experience. It provides functionality to list namespaces, pods, containers, and files within a pod's container. It also supports file download and upload operations between the local environment and containers in a Kubernetes pod.

![podFiles](podFiles.png)

## Install

```sh
# 1. Initialize a sparse checkout repository
git clone --filter=blob:none --sparse https://github.com/zrcoder/podFiles
cd podFiles

# 2. Set to download only deploy directory
git sparse-checkout set cmd/deploy

# 3. Check downloaded files
ls -la cmd/deploy

# 4. Run the install script
cd cmd/deploy
chmod +x apply.sh

# 5. Deploy with custom settings (optional)
./apply.sh namespace=my-ns image=my-registry.com/podfiles:v1.0.0 domain=pods.example.com

# Or use default settings
./apply.sh

# 6. Verify the deployment
kubectl get pods -n <namespace>
```

After successful deployment, you can access PodFiles through:

- If ingress is enabled: https://your-domain
