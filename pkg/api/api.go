package api

import (
	"archive/tar"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/zrcoder/podFiles/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/schema"
	"github.com/zrcoder/podFiles/internal/k8s"
)

const (
	Prefix = "/api/"

	loginPath      = "login"
	registerPath   = "register"
	logoutPath     = "logout"
	unregisterPath = "unregister"
	userPath       = "user"

	namespacesPath = "namespaces"
	podsPath       = "pods"
	containersPath = "containers"
	filesPath      = "files"
	uploadPath     = "upload"
	downloadPath   = "download"

	Login      = Prefix + loginPath
	Register   = Prefix + registerPath
	Logout     = Prefix + logoutPath
	Unregister = Prefix + unregisterPath
	User       = Prefix + userPath

	Namespaces = Prefix + namespacesPath
	Pods       = Prefix + podsPath
	Containers = Prefix + containersPath
	Files      = Prefix + filesPath
	Upload     = Prefix + uploadPath
	Download   = Prefix + downloadPath
)

var k8sClient *k8s.Client

var (
	statemux sync.RWMutex
	state    = &models.State{}
)

func New() http.Handler {
	gin.SetMode(gin.ReleaseMode)

	var err error
	k8sClient, err = k8s.New()
	if err != nil {
		panic(err)
	}

	handler := gin.Default()
	api := handler.Group(Prefix)
	{
		api.GET(namespacesPath, listNamespaces)
		api.POST(namespacesPath, setNamespace)
		api.GET(podsPath, listPods)
		api.POST(podsPath, setPod)
		api.GET(containersPath, listContainers)
		api.POST(containersPath, setContainer)
		api.GET(filesPath, listFiles)
		api.POST(uploadPath, upload)
		api.POST(downloadPath, download)
	}
	return handler
}

func listNamespaces(c *gin.Context) {
	ns, err := k8sClient.ListNamespaces(c.Request.Context())
	if err != nil {
		slog.Error("list namespaces", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schema.SuccessResponse("", schema.Schema{"items": ns, "total": len(ns)}))
}

func setNamespace(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		slog.Error("namespace is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}
	statemux.Lock()
	state.SetNamespace(namespace)
	statemux.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": ""})
}

func listPods(c *gin.Context) {
	statemux.RLock()
	defer statemux.RUnlock()
	if state.Namespace == "" {
		slog.Info("namespace is required")
		c.JSON(http.StatusOK, []models.Pod{})
		return
	}
	pods, err := k8sClient.ListPods(c.Request.Context(), state.Namespace)
	if err != nil {
		slog.Error("list pods", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pods)
}

func setPod(c *gin.Context) {
	pod := c.Query("pod")
	if pod == "" {
		slog.Error("pod is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pod is required"})
		return
	}
	statemux.Lock()
	state.SetPod(pod)
	statemux.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": ""})
}

func listContainers(c *gin.Context) {
	statemux.RLock()
	defer statemux.RUnlock()
	if state.Namespace == "" || state.Pod == "" {
		slog.Info("namespace or pod is required")
		c.JSON(http.StatusOK, []models.Container{})
		return
	}
	containers, err := k8sClient.ListContainers(c.Request.Context(), state.Namespace, state.Pod)
	if err != nil {
		slog.Error("list containers", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func setContainer(c *gin.Context) {
	container := c.Query("container")
	if container == "" {
		slog.Error("container is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "container is required"})
		return
	}
	statemux.Lock()
	state.SetContainer(container)
	statemux.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": ""})
}

func listFiles(c *gin.Context) {
	statemux.RLock()
	defer statemux.RUnlock()
	if state.Namespace == "" || state.Pod == "" || state.Container == "" {
		slog.Info("namespace, pod or container is required")
		c.JSON(http.StatusOK, []models.FileInfo{})
		return
	}
	files, err := k8sClient.ListFiles(c.Request.Context(), state.Namespace, state.Pod, state.Container, state.FSPath())
	if err != nil {
		slog.Error("list files", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

func upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		slog.Error("upload file", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		slog.Error("upload file", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		tw := tar.NewWriter(pw)
		defer tw.Close()

		hdr := &tar.Header{
			Name: file.Filename,
			Mode: 0o644,
			Size: file.Size,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			slog.Error("upload file", "err", err)
			return
		}
		if _, err := io.Copy(tw, src); err != nil {
			slog.Error("upload file", "err", err)
			return
		}
	}()
	statemux.RLock()
	defer statemux.RUnlock()
	err = k8sClient.UploadFile(c.Request.Context(), state.Namespace, state.Pod, state.Container, state.FSPath(), pr)
	if err != nil {
		slog.Error("upload file", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func download(c *gin.Context) {
	statemux.RLock()
	defer statemux.RUnlock()
	file := c.Query("file")
	path := state.FSPath() + "/" + file
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(path)))

	err := k8sClient.DownloadFile(c.Request.Context(), state.Namespace, state.Pod, state.Container, path, c.Writer)
	if err != nil {
		slog.Error("download file", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
