package main

import "strings"

type StringArtifact struct {
	sb strings.Builder
}

func (sa *StringArtifact) BuilderId() string {
	return "osbuild.image-builder"
}

func (sa *StringArtifact) Files() []string {
	return []string{}
}

func (sa *StringArtifact) Id() string {
	return ""
}

func (sa *StringArtifact) String() string {
	return sa.sb.String()
}

func (sa *StringArtifact) State(name string) interface{} {
	return nil
}

func (sa *StringArtifact) Destroy() error {
	return nil
}

func (sa *StringArtifact) WriteString(s string) {
	sa.sb.WriteString(s)
}
