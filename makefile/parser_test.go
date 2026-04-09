package makefile

import (
	"strings"
	"testing"
)

func TestParseReader_Basic(t *testing.T) {
	src := `
APP_NAME = myapp
VERSION ?= 1.0.0
ENV := development
GIT_SHA != git rev-parse HEAD

## build: アプリをビルドする
build:
	go build -o $(APP_NAME) -ldflags "-X main.Version=$(VERSION)" ./...

## 開発用ビルド
build-dev:
	go build -tags dev -o $(APP_NAME)-dev ./...

.PHONY: build build-dev

rebuild: build
	@echo "$(APP_NAME) rebuilt"
`
	pm, err := parseReader(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if got, want := len(pm.Targets), 3; got != want {
		t.Fatalf("want %d targets, got %d: %+v", want, got, pm.Targets)
	}

	checkTarget := func(idx int, name, desc string, vars []string) {
		t.Helper()
		tg := pm.Targets[idx]
		if tg.Name != name {
			t.Errorf("target %d: name=%q, want %q", idx, tg.Name, name)
		}
		if tg.Description != desc {
			t.Errorf("target %d: desc=%q, want %q", idx, tg.Description, desc)
		}
		if len(tg.Variables) != len(vars) {
			t.Errorf("target %d: vars=%v, want %v", idx, tg.Variables, vars)
			return
		}
		for i, v := range vars {
			if tg.Variables[i] != v {
				t.Errorf("target %d: var[%d]=%q, want %q", idx, i, tg.Variables[i], v)
			}
		}
	}

	checkTarget(0, "build", "アプリをビルドする", []string{"APP_NAME", "VERSION"})
	checkTarget(1, "build-dev", "開発用ビルド", []string{"APP_NAME"})
	checkTarget(2, "rebuild", "", []string{"APP_NAME"})

	wantDefaults := map[string]string{
		"APP_NAME": "myapp",
		"VERSION":  "1.0.0",
		"ENV":      "development",
		"GIT_SHA":  "",
	}
	for k, v := range wantDefaults {
		if got := pm.Defaults[k]; got != v {
			t.Errorf("default %s=%q, want %q", k, got, v)
		}
	}
}

func TestParseReader_SkipsDotTargets(t *testing.T) {
	src := `
.PHONY: all
all:
	@echo all
`
	pm, err := parseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(pm.Targets) != 1 || pm.Targets[0].Name != "all" {
		t.Errorf("unexpected targets: %+v", pm.Targets)
	}
}

func TestParseReader_FirstDefinitionWins(t *testing.T) {
	src := `
build:
	@echo first

build:
	@echo second
`
	pm, err := parseReader(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(pm.Targets) != 1 {
		t.Fatalf("want 1 target, got %d", len(pm.Targets))
	}
	if pm.Targets[0].LineNumber != 2 {
		t.Errorf("want first definition, got line %d", pm.Targets[0].LineNumber)
	}
}

func TestStripInlineComment(t *testing.T) {
	cases := map[string]string{
		"myapp":                "myapp",
		"myapp # the app":      "myapp",
		"my\\#app # the app":   "my\\#app",
		"":                     "",
		"  value  # comment  ": "  value",
	}
	for in, want := range cases {
		if got := stripInlineComment(in); got != want {
			t.Errorf("stripInlineComment(%q)=%q, want %q", in, got, want)
		}
	}
}
