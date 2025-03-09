package api

import (
	"archive/tar"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zrcoder/amisgo/schema"
	"github.com/zrcoder/podFiles/internal/k8s"
	"github.com/zrcoder/podFiles/pkg/auth"
	"github.com/zrcoder/podFiles/pkg/models"
	"github.com/zrcoder/podFiles/pkg/state"
	"github.com/zrcoder/podFiles/pkg/util/log"
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
	g.Use(auth.Auth)
	api := g.Group(Prefix)
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
	ns, err := k8sClient.ListNamespaces(c.Request.Context())
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
	pods, err := k8sClient.ListPods(c.Request.Context(), session)
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
	files, err := k8sClient.ListFiles(c.Request.Context(), c.GetString(state.SessionKey))
	if err != nil {
		slog.Error("list files", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
	slog.Debug("list files", "files", files)
	c.JSON(http.StatusOK, schema.SuccessResponse("success", schema.Schema{
		"files":      files,
		"breadItems": getBreadcrumbs(st),
		"inSubDir":   st.InSubDir(),
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

func getBreadcrumbs(st *models.State) []models.BreadcrumbItem {
	return []models.BreadcrumbItem{
		{Label: st.Namespace},
		{Label: st.Pod},
		{Label: st.Container},
		{Label: st.FSPath()},
	}
}

func upload(c *gin.Context) {
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
			slog.Error("upload file", log.Error(err))
			return
		}
		if _, err := io.Copy(tw, src); err != nil {
			slog.Error("upload file", log.Error(err))
			return
		}
	}()
	session := c.GetString(state.SessionKey)
	st := state.Get(session)
	err = k8sClient.UploadFile(c.Request.Context(), st.Namespace, st.Pod, st.Container, st.FSPath(), pr)
	if err != nil {
		slog.Error("upload file", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
}

func download(c *gin.Context) {
	session := c.GetString(state.SessionKey)
	file := c.Query("file")
	file = strings.Trim(file, "/")
	slog.Debug("download", slog.String("path", file))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tgz", file))
	err := k8sClient.DownloadFile(c.Request.Context(), session, file, c.Writer)
	if err != nil {
		slog.Error("download file failed", log.Error(err))
		c.JSON(http.StatusInternalServerError, schema.ErrorResponse(err.Error()))
		return
	}
}
