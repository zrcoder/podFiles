package api

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/schema"
	"github.com/zrcoder/podFiles/internal/auth"
	"github.com/zrcoder/podFiles/internal/k8s"
	"github.com/zrcoder/podFiles/internal/models"
	"github.com/zrcoder/podFiles/internal/state"
	"github.com/zrcoder/podFiles/internal/util/log"
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
	fsPathPath     = "fsPath"
	uploadPath     = "upload"
	downloadPath   = "download"

	HealthPath = "/health"

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

func New() http.Handler {
	gin.SetMode(gin.ReleaseMode)

	var err error
	k8sClient, err = k8s.New()
	if err != nil {
		panic(err)
	}

	g := gin.Default()
	api := g.Group(Prefix)
	api.Use(auth.Auth)
	{
		api.GET(namespacesPath, listNamespaces)
		api.POST(namespacesPath, setNamespace)
		api.GET(podsPath, listPods)
		api.POST(podsPath, setPod)
		api.GET(containersPath, listContainers)
		api.POST(containersPath, setContainer)
		api.GET(filesPath, listFiles)
		api.POST(filesPath, setPath)
		api.POST(uploadPath, upload)
		api.POST(downloadPath, download)
	}

	return g
}

func listNamespaces(c *gin.Context) {
	ns, err := k8sClient.ListCommonNamespaces(c.Request.Context())
	if err != nil {
		slog.Error("list namespaces", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, ns)
}

func setNamespace(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		slog.Error("namespace is required")
		c.JSON(http.StatusBadRequest, schema.ErrorResponse("namespace is required"))
		return
	}
	session := c.GetString(state.SessionKey)
	state.Get(session).SetNamespace(namespace)
	c.Status(http.StatusOK)
}

func listPods(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	slog.Debug("list pods", slog.String("session", session))
	pods, err := k8sClient.ListRunningPods(c.Request.Context(), session)
	if err != nil {
		slog.Error("list pods", log.Error(err))
		c.JSON(http.StatusOK, []models.Pod{})
		return
	}
	c.JSON(http.StatusOK, pods)
}

func setPod(c *gin.Context) {
	pod := c.Query("pod")
	if pod == "" {
		slog.Error("pod is required")
		c.JSON(http.StatusBadRequest, schema.ErrorResponse("pod is required"))
		return
	}
	session := c.GetString(state.SessionKey)
	state.Get(session).SetPod(pod)
	c.Status(http.StatusOK)
}

func listContainers(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	containers, err := k8sClient.ListContainers(c.Request.Context(), session)
	if err != nil {
		slog.Error("list containers", log.Error(err))
		c.JSON(http.StatusOK, []models.Container{})
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
	session := c.GetString(state.SessionKey)
	state.Get(session).SetContainer(container)
	c.Status(http.StatusOK)
}

func listFiles(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	st := state.Get(session)
	if st == nil || st.Container == "" {
		c.JSON(http.StatusBadRequest, schema.ErrorResponse("container is required"))
		return
	}
	files, err := k8sClient.ListFiles(c.Request.Context(), st)
	if err != nil {
		slog.Error("list files", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
	slog.Debug("list files", "files", files)
	c.JSON(http.StatusOK, schema.SuccessResponse("success", schema.Schema{
		"files": files,
		"breadItems": []models.BreadcrumbItem{
			{Label: st.Namespace},
			{Label: st.Pod},
			{Label: st.Container},
			{Label: st.FSPath()},
		},
		"inSubDir": st.InSubDir(),
	}))
}

func setPath(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	st := state.Get(session)
	back := c.Query("back")
	if back == "true" {
		popPath(c, st)
		return
	}
	appendPath(c, st)
}

func popPath(c *gin.Context, st *models.State) {
	err := st.PopPath()
	if err != nil {
		slog.Error("pop path", log.Error(err))
		c.JSON(http.StatusBadRequest, schema.ErrorResponse(err.Error()))
		return
	}
}

func appendPath(c *gin.Context, st *models.State) {
	dir := strings.TrimRight(c.Query("dir"), "/")
	if dir == "" {
		c.JSON(http.StatusBadRequest, schema.ErrorResponse("directory is required"))
		return
	}
	slog.Debug("append path", slog.String("path", dir))

	st.AddPath(dir)
	c.Status(http.StatusOK)
}

func upload(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		slog.Error("upload file", log.Error(err))
		c.JSON(http.StatusBadRequest, schema.ErrorResponse(err.Error()))
		return
	}

	src, err := file.Open()
	if err != nil {
		slog.Error("upload file", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
	defer src.Close()

	// Create a pipe for streaming data
	pr, pw := io.Pipe()

	// Use a goroutine to write data to the pipe
	go func() {
		defer pw.Close()

		// Use buffered writer to reduce memory pressure
		bufWriter := bufio.NewWriterSize(pw, k8s.FileBufferSize)
		tw := tar.NewWriter(bufWriter)
		defer tw.Close()

		hdr := &tar.Header{
			Name: file.Filename,
			Mode: 0o644,
			Size: file.Size,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			slog.Error("failed to write tar header", log.Error(err))
			return
		}

		// Use a buffer for copying to control memory usage
		buf := make([]byte, k8s.FileBufferSize)

		// Use CopyBuffer instead of Copy for better memory control
		_, err := io.CopyBuffer(tw, src, buf)
		if err != nil {
			slog.Error("failed to copy file data", log.Error(err))
			return
		}

		// Ensure all data is flushed
		if err := tw.Flush(); err != nil {
			slog.Error("failed to flush tar writer", log.Error(err))
			return
		}

		if err := bufWriter.Flush(); err != nil {
			slog.Error("failed to flush buffer", log.Error(err))
			return
		}
	}()

	session := c.GetString(state.SessionKey)
	st := state.Get(session)

	// Upload the file to the pod
	err = k8sClient.UploadFile(c.Request.Context(), st.Namespace, st.Pod, st.Container, st.FSPath(), pr)
	if err != nil {
		slog.Error("upload file", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, schema.SuccessResponse("", schema.Schema{"value": "success"}))
}

func download(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	file := c.Query("file")
	file = strings.Trim(file, "/")
	slog.Debug("download", slog.String("path", file))

	// Set response headers
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tgz", file))

	// Set Transfer-Encoding to chunked for streaming
	c.Header("Transfer-Encoding", "chunked")

	// Use Gin's Stream method for streaming response
	c.Stream(func(w io.Writer) bool {
		// Create a buffered writer to reduce memory pressure
		bufWriter := bufio.NewWriterSize(w, k8s.FileBufferSize)

		err := k8sClient.DownloadFile(c.Request.Context(), session, file, bufWriter)
		if err != nil {
			slog.Error("download file failed", log.Error(err))
		}

		// Ensure all data is flushed
		bufWriter.Flush()

		return false // Return false to end the stream
	})
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
