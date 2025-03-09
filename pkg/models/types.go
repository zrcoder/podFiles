package models

import (
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

// func (s *State) Normalize() {
// 	s.ApiPath = s.Api()
// 	s.FSPath = s.fsPath()
// }

// func (s *State) Api() string {
// 	path := []string{s.Namespace, s.Pod, s.Container}
// 	fsPath := s.fsPath()
// 	if fsPath != "" {
// 		path = append(path, fsPath)
// 	}
// 	return strings.Join(path, "/")
// }

func (s *State) FSPath() string {
	return strings.Join(s.Path, "/")
}
