package k8s

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/zrcoder/podFiles/pkg/models"
	"github.com/zrcoder/podFiles/pkg/state"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// FileBufferSize defines the standard buffer size used for I/O operations
const FileBufferSize = 32 * 1024 // 32KB

// Client represents a Kubernetes client
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

func (c *Client) ListPods(ctx context.Context, session string) ([]models.Pod, error) {
	st := state.Get(session)
	if st.Namespace == "" {
		msg := "namespace is required"
		slog.Error(msg)
		return nil, errors.New(msg)
	}
	list, err := c.clientset.CoreV1().Pods(st.Namespace).List(ctx, metav1.ListOptions{})
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

func (c *Client) ListContainers(ctx context.Context, session string) ([]models.Container, error) {
	st := state.Get(session)
	if st.Namespace == "" || st.Pod == "" {
		msg := "namespace and pod are required"
		slog.Error(msg)
		return nil, errors.New(msg)
	}
	p, err := c.clientset.CoreV1().Pods(st.Namespace).Get(ctx, st.Pod, metav1.GetOptions{})
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

func (c *Client) ListFiles(ctx context.Context, st *models.State) ([]models.FileInfo, error) {
	if st.Namespace == "" || st.Pod == "" || st.Container == "" {
		msg := "namespace, pod or container is required"
		slog.Info(msg)
		return nil, errors.New(msg)
	}

	dir := st.FSPath()
	if dir == "" {
		dir = "/"
	}

	slog.Debug("list files", "namespace", st.Namespace, "pod", st.Pod, "container", st.Container, "dir", dir)

	cmd := []string{"ls", "-lh", dir}
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(st.Pod).Namespace(st.Namespace).SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: st.Container,
			Command:   cmd,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

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
	fileTypeMap := map[string]string{
		"d": "dir",
		"-": "file",
		"l": "link",
	}
	files := []models.FileInfo{}
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
		fileType, ok := fileTypeMap[fields[0][0:1]]
		if !ok {
			continue
		}
		name := strings.Join(fields[8:], " ")
		file := models.FileInfo{
			Name: name,
			Type: fileType,
			Size: fields[4],
			Time: fields[5] + "-" + fields[6] + " " + fields[7],
		}
		files = append(files, file)
	}
	return files
}

func (c *Client) DownloadFile(ctx context.Context, session, filePath string, writer io.Writer) error {
	st := state.Get(session)
	if st.Namespace == "" || st.Pod == "" || st.Container == "" {
		return errors.New("namespace, pod or container is required")
	}

	// Use shell command to execute tar
	shellCmd := fmt.Sprintf("cd %s && tar czf - %s", st.FSPath(), filePath)
	cmd := []string{"/bin/sh", "-c", shellCmd}

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(st.Pod).Namespace(st.Namespace).SubResource("exec").
		Param("container", st.Container).
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdout:  true,
			Stderr:  true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Create a buffer for stderr
	errBuf := new(bytes.Buffer)

	// Stream directly to the writer without buffering the entire content in memory
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: writer,
		Stderr: errBuf,
	})
	if err != nil {
		errMsg := err.Error()
		if errBuf.Len() > 0 {
			errMsg = fmt.Sprintf("%v: %s", err, errBuf.String())
		}
		return fmt.Errorf("exec error: %v", errMsg)
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, namespace, pod, container, targetDir string, reader io.Reader) error {
	// Use buffered reader to control memory usage
	bufReader := bufio.NewReaderSize(reader, FileBufferSize)

	cmd := []string{"tar", "xf", "-", "-C", targetDir}
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").Name(pod).Namespace(namespace).SubResource("exec").
		Param("container", container).
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	errBuf := new(bytes.Buffer)

	// Stream data directly from reader to pod without buffering entire content
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  bufReader,
		Stdout: io.Discard, // Discard stdout
		Stderr: errBuf,
	})
	if err != nil {
		errMsg := err.Error()
		if errBuf.Len() > 0 {
			errMsg = fmt.Sprintf("%v: %s", err, errBuf.String())
		}
		return fmt.Errorf("exec error: %v", errMsg)
	}

	return nil
}
