package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/zrcoder/podFiles/pkg/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func New() (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{clientset: clientset, config: config}, nil
}

func (c *Client) ListNamespaces(ctx context.Context) ([]models.Namespace, error) {
	list, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	ns := make([]models.Namespace, 0, len(list.Items))
	for _, n := range list.Items {
		slog.Debug("namespace", "name", n.Name)
		ns = append(ns, models.Namespace{
			Namespace: n.Name,
		})
	}
	return ns, nil
}

func (c *Client) ListPods(ctx context.Context, namespace string) ([]models.Pod, error) {
	list, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]models.Pod, 0, len(list.Items))
	for _, pod := range list.Items {
		slog.Debug("pod", "name", pod.Name)
		names = append(names, models.Pod{
			Pod: pod.Name,
		})
	}
	return names, nil
}

func (c *Client) ListContainers(ctx context.Context, namespace, pod string) ([]models.Container, error) {
	p, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, pod, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	containers := make([]models.Container, 0, len(p.Spec.Containers))
	for _, c := range p.Spec.Containers {
		slog.Debug("container", "name", c.Name)
		containers = append(containers, models.Container{
			Container: c.Name,
		})
	}
	return containers, nil
}

func (c *Client) ListFiles(ctx context.Context, namespace, pod, container, dir string) ([]models.FileInfo, error) {
	if dir == "" {
		dir = "."
	}
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").
		Param("container", container).
		Param("command", "ls").
		Param("command", "-lhF").
		Param("command", dir).
		Param("stdout", "true").
		Param("stderr", "true")

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return nil, err
	}

	output := bytes.NewBuffer(nil)
	outputErr := bytes.NewBuffer(nil)
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: output,
		Stderr: outputErr,
	})
	if err != nil {
		return nil, fmt.Errorf("exec error: %w, output: %s", err, output.String())
	}
	if outputErr.Len() > 0 {
		return nil, fmt.Errorf("exec error: %s", outputErr.String())
	}

	return parseFileList(output.String()), nil
}

// parseFileList parses the output of the `ls -lF` command and returns a slice of FileInfo structs.
func parseFileList(output string) []models.FileInfo {
	var files []models.FileInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Skip empty lines and lines starting with "total"
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}
		fields := strings.Fields(line)
		// Skip lines with insufficient fields
		if len(fields) < 9 {
			continue
		}
		name := fields[8]
		isDir := strings.HasSuffix(name, "/")
		file := models.FileInfo{
			Name:  name,
			IsDir: isDir,
			Size:  fields[4],
			Time:  fields[5] + "-" + fields[6] + " " + fields[7],
		}
		files = append(files, file)
	}
	return files
}

func (c *Client) DownloadFile(ctx context.Context, namespace, pod, container, filePath string, writer io.Writer) error {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").
		Param("container", container).
		Param("command", fmt.Sprintf("tar cf -C %s", filePath))
	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: writer,
		Stderr: os.Stderr,
	})
}

func (c *Client) UploadFile(ctx context.Context, namespace, pod, container, targetDir string, reader io.Reader) error {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").
		Param("container", container).
		Param("command", fmt.Sprintf("tar xf -C %s", targetDir))

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	return executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
