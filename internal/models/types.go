package models

import (
	"errors"
	"strings"
)

type Namespace struct {
	Namespace string `json:"namespace"`
}

type Pod struct {
	Pod string `json:"pod"`
}

type Container struct {
	Container string `json:"container"`
}

type BreadcrumbItem struct {
	Label string `json:"label"`
}

type FileInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size string `json:"size"`
	Time string `json:"time"`
}

type State struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container"`
	Path      []string `json:"path"`
}

func (s *State) SetNamespace(namespace string) {
	s.Namespace = namespace
	s.SetPod("")
}

func (s *State) SetPod(pod string) {
	s.Pod = pod
	s.SetContainer("")
}

func (s *State) SetContainer(container string) {
	s.Container = container
	s.SetPath(nil)
}

func (s *State) SetPath(path []string) {
	s.Path = path
}

func (s *State) AddPath(path string) {
	s.Path = append(s.Path, path)
}

func (s *State) PopPath() error {
	if len(s.Path) == 0 {
		return errors.New("path is empty")
	}
	s.Path = s.Path[:len(s.Path)-1]
	return nil
}

func (s *State) FSPath() string {
	return "/" + strings.Join(s.Path, "/")
}

func (s *State) InSubDir() bool {
	return len(s.Path) > 0
}
