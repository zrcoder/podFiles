# PodFiles

PodFiles is a tool based on [amisgo](https://github.com/zrcoder/amisgo), designed to manage files within Kubernetes pods.

It provides a user-friendly interface to list namespaces, pods, containers, as well as directories and files within the containers, and supports file uploads and downloads between your local environment and pod containers.

![podFiles](podFiles.png)

## Installation

You can install PodFiles using one of the following methods:

### 1. **Deploy as a Pod in the Cluster**

Use the interactive script below:

```sh
# 1. Download the deployment script
wget https://raw.githubusercontent.com/zrcoder/podFiles/main/cmd/deploy/apply.sh

# 2. Make the script executable
chmod +x apply.sh

# 3. Run the deployment script
./apply.sh

# 4. Follow the prompts to configure your deployment.
# The script will guide you through setting the namespace, image, and so on.
```

### 2. **Install the PodFiles Binary**

This can be deployed inside or outside the cluster, requiring a kubeconfig file.

```sh
go install github.com/zrcoder/podFiles/cmd/podFiles@latest

# Assuming the kubeconfig file is ~/.kube/config
KUBECONFIG=~/.kube/config nohup podFiles > podFiles.log 2>&1 &
```

> Use the _NS_BLACK_LIST_ environment variable to specify namespaces to ignore.
>
> ```sh
> KUBECONFIG=~/.kube/config NS_BLACK_LIST=kube-,default nohup podFiles > podFiles.log 2>&1 &
> ```
>
> By default, PodFiles uses port 8080. You can specify a different port with the _PORT_ environment variable:
>
> ```sh
> KUBECONFIG=~/.kube/config PORT=8081 nohup podFiles > podFiles.log 2>&1 &
> ```
