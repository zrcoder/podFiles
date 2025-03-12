# PodFiles

PodFiles is a Go - based project designed to interact with Kubernetes clusters. It features a concise UI that simplifies the user experience. It provides functionality to list namespaces, pods, containers, and files within a pod's container. It also supports file download and upload operations between the local environment and containers in a Kubernetes pod.

![podFiles](podFiles.png)

## Install

```sh
# 1. Download the deployment script
wget https://raw.githubusercontent.com/zrcoder/podFiles/main/cmd/deploy/apply.sh

# 2. Make the script executable
chmod +x apply.sh

# 3. Run the deployment script
./apply.sh

# 4. Follow the interactive prompts to configure your deployment settings.
# The script will guide you through setting the namespace, image and so on.
```
