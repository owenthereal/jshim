package jshim

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

type Env map[string]string

func New(args ...string) *JShim {
	var name string
	if runtime.GOOS == "windows" {
		name = "java.exe"
	} else {
		name = "java"
	}

	binary, err := lookupJavaBin(name)
	if err != nil {
		fatal(err)
	}

	args = append(args, os.Args[1:]...)
	return &JShim{
		Binary: binary,
		Args:   args,
		Env:    make(map[string]string),
	}
}

type JShim struct {
	Binary string
	Args   []string
	Env    Env
}

func (j *JShim) Execute() {
	for k, v := range j.Env {
		os.Setenv(k, v)
	}

	var err error
	if runtime.GOOS == "windows" {
		err = j.spawnCmd()
	} else {
		err = j.execCmd()
	}

	if err != nil {
		fatal(err)
	}
}

func (j *JShim) spawnCmd() error {
	cmd := exec.Command(j.Binary, j.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (j *JShim) execCmd() error {
	binary, err := exec.LookPath(j.Binary)
	if err != nil {
		return err
	}

	args := append([]string{j.Binary}, j.Args...)
	env := os.Environ()
	return syscall.Exec(binary, args, env)
}

func lookupJavaBin(name string) (string, error) {
	pathJavaBin, err := exec.LookPath(name)
	if err == nil {
		return pathJavaBin, nil
	}

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		ibmJavaBin := filepath.Join(javaHome, "jre", "sh", name)
		if _, e := os.Stat(ibmJavaBin); e == nil {
			return ibmJavaBin, nil
		}

		commonJavaBin := filepath.Join(javaHome, "bin", name)
		if _, e := os.Stat(commonJavaBin); e == nil {
			return commonJavaBin, nil
		}
	}

	return "", fmt.Errorf("JAVA_HOME is not set and no 'java' command could be found in your PATH.")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
