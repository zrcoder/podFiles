# PodFiles

PodFiles is a Go-based tool designed for managing files within Kubernetes pods. It provides a user-friendly interface to list namespaces, pods, and containers, and supports file download and upload operations between the local environment and pod containers.

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
