package boshrelease

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Repository struct {
	repository string
	branch     string
	tmpdir     string
	privateKey string
}

func NewRepository(repository, branch, privateKey string) *Repository {
	cs := sha1.New()
	cs.Write([]byte(repository))
	cs.Write([]byte(branch))

	return &Repository{
		repository: repository,
		branch:     branch,
		privateKey: privateKey,
		tmpdir:     path.Join(os.TempDir(), fmt.Sprintf("bosh-release-%x", cs.Sum(nil))),
	}
}

func (r Repository) Path() string {
	return r.tmpdir
}

func (r Repository) Pull() error {
	var args []string

	if _, err := os.Stat(path.Join(r.tmpdir, ".git")); os.IsNotExist(err) {
		args = []string{"clone", "--quiet", r.repository}

		if r.branch != "" {
			args = append(args, "--branch", r.branch)
		}

		args = append(args, ".")

		err = os.MkdirAll(r.tmpdir, 0700)
		if err != nil {
			return errors.Wrap(err, "mkdir local repo")
		}
	} else {
		args = []string{"pull", "--ff-only", "--quiet", r.repository}

		if r.branch != "" {
			args = append(args, r.branch)
		}
	}

	err := r.run(args...)
	if err != nil {
		return errors.Wrap(err, "fetching repository")
	}

	// TODO reset to handle force push?

	return nil
}

func (r Repository) Push() error {
	return errors.New("TODO")
}

func (r Repository) Configure(authorName, authorEmail string) error {
	configs := map[string]string{
		"user.name":  authorName,
		"user.email": authorEmail,
	}

	for k, v := range configs {
		err := r.run("config", k, v)
		if err != nil {
			return errors.Wrapf(err, "setting %s", k)
		}
	}

	return nil
}

func (r Repository) Commit(message string, rebase bool) (string, error) {
	err := r.run("add", "-A", ".")
	if err != nil {
		return "", errors.Wrap(err, "adding files")
	}

	err = r.run("commit", "-m", message)
	if err != nil {
		return "", errors.Wrap(err, "committing")
	}

	attempts := 0
	if rebase {
		attempts = 3
	}

	var finalError error

	for true {
		err := r.run("push", "origin", fmt.Sprintf("HEAD:%s", r.branch))
		if err == nil {
			break
		}

		if attempts <= 0 {
			finalError = err

			break
		}

		time.Sleep(5 * time.Second)

		err = r.run("pull", "--rebase", r.repository, r.branch)
		if err != nil {
			return "", errors.Wrap(err, "rebasing")
		}

		err = r.run("commit", "--amend", "--reset-author", "--no-edit")
		if err != nil {
			return "", errors.Wrap(err, "resetting commit")
		}

		attempts--
	}

	if finalError != nil {
		return "", finalError
	}

	stdout := bytes.NewBuffer(nil)

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.tmpdir
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "resolving HEAD")
	}

	return strings.TrimSpace(stdout.String()), nil
}

func (r Repository) Tag(commit, tag, message string) error {
	err := r.run("tag", "-a", "-m", message, tag, commit)
	if err != nil {
		return errors.Wrap(err, "tagging")
	}

	err = r.run("push", "origin", tag)
	if err != nil {
		return errors.Wrap(err, "pushing tag")
	}

	return nil
}

func (r Repository) run(args ...string) error {
	var executable = "git"

	if r.privateKey != "" && (args[0] == "clone" || args[0] == "pull" || args[0] == "push") {
		privateKey, err := ioutil.TempFile("", "git-privateKey")
		if err != nil {
			return errors.Wrap(err, "tempfile for id_rsa")
		}

		defer os.RemoveAll(privateKey.Name())

		err = os.Chmod(privateKey.Name(), 0600)
		if err != nil {
			return errors.Wrap(err, "chmod git wrapper")
		}

		err = ioutil.WriteFile(privateKey.Name(), []byte(r.privateKey), 0600)
		if err != nil {
			return errors.Wrap(err, "writing id_rsa")
		}

		executableWrapper, err := ioutil.TempFile("", "git-executable")
		if err != nil {
			return errors.Wrap(err, "tempfile for git wrapper")
		}

		defer os.RemoveAll(executableWrapper.Name())

		err = ioutil.WriteFile(executableWrapper.Name(), []byte(fmt.Sprintf(`#!/bin/bash

set -eu

eval $(ssh-agent) > /dev/null

trap "kill $SSH_AGENT_PID" 0

SSH_ASKPASS=false DISPLAY= ssh-add "%s" 2>/dev/null # TODO suppresses real errors?

exec git "$@"`, privateKey.Name())), 0500)
		if err != nil {
			return errors.Wrap(err, "writing git wrapper")
		}

		err = os.Chmod(executableWrapper.Name(), 0500)
		if err != nil {
			return errors.Wrap(err, "chmod git wrapper")
		}

		executable = executableWrapper.Name()
	}

	cmd := exec.Command(executable, args...)
	cmd.Dir = r.tmpdir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
