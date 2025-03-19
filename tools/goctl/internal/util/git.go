package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/env"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
)

func CloneIntoGitHome(url, branch string) (dir string, err error) {
	gitHome, err := pathx.GetGitHome()
	if err != nil {
		return "", err
	}
	os.RemoveAll(gitHome)
	ext := filepath.Ext(url)
	repo := strings.TrimSuffix(filepath.Base(url), ext)
	dir = filepath.Join(gitHome, repo)
	if pathx.FileExists(dir) {
		os.RemoveAll(dir)
	}
	path, err := env.LookPath("git")
	if err != nil {
		return "", err
	}

	args := []string{"clone"}
	if len(branch) > 0 {
		args = append(args, "-b", branch)
	}
	args = append(args, url, dir)
	cmd := exec.Command(path, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}
