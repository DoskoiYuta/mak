package mkfile

import (
	"reflect"
	"testing"
)

func TestBuildArgs(t *testing.T) {
	got := BuildArgs(RunOptions{
		Makefile:  "./Makefile",
		Directory: "sub",
		Target:    "build",
		Variables: []VarAssignment{
			{"APP_NAME", "myapp"},
			{"VERSION", "1.0.0"},
		},
	})
	want := []string{"make", "-f", "./Makefile", "-C", "sub", "build", "APP_NAME=myapp", "VERSION=1.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs = %v, want %v", got, want)
	}
}

func TestBuildArgs_OmitsEmpty(t *testing.T) {
	got := BuildArgs(RunOptions{Target: "build"})
	want := []string{"make", "build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildArgs = %v, want %v", got, want)
	}
}

func TestPreviewCommand_Quoting(t *testing.T) {
	got := PreviewCommand(RunOptions{
		Target: "build",
		Variables: []VarAssignment{
			{"MSG", "hello world"},
		},
	})
	want := "$ make build 'MSG=hello world'"
	if got != want {
		t.Errorf("PreviewCommand = %q, want %q", got, want)
	}
}
