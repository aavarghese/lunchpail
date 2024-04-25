package demo

import (
	"embed"
	"io/ioutil"
	"os"

	"lunchpail.io/pkg/shrinkwrap"
)

//go:generate /bin/bash -c "tar --exclude '*.md' -zcf demo.tar.gz -C ../../tests/tests/test7f-by-role-autorun-autodispatcher/pail ."
//go:embed demo.tar.gz
var demo embed.FS

func expand() (string, error) {
	if dir, err := ioutil.TempDir("", "lunchpail"); err != nil {
		return "", err
	} else if err := shrinkwrap.Expand(dir, demo, "demo.tar.gz", true); err != nil {
		return "", err
	} else {
		return dir, nil
	}
}

func Shrinkwrap(opts Options) error {
	demoPath, err := expand()
	if err != nil {
		return err
	}

	appName := "demo"
	clusterIsOpenShift := false
	workdirViaMount := false
	imagePullSecret := ""
	branch := ""
	overrideValues := []string{}
	queue := ""
	needsCsiH3 := false
	needsCsiS3 := false
	needsCsiNfs := false
	hasGpuSupport := false
	dockerHost := ""

	defer os.RemoveAll(demoPath)
	return shrinkwrap.App(
		demoPath,
		opts.OutputDir,
		shrinkwrap.AppOptions{opts.Namespace, appName, clusterIsOpenShift, workdirViaMount, imagePullSecret, branch, overrideValues, opts.Verbose, queue, needsCsiH3, needsCsiS3, needsCsiNfs, hasGpuSupport, dockerHost, opts.Force},
	)
}
