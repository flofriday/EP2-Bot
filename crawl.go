package main

import (
	"gopkg.in/src-d/go-git.v4"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
)

// Clone the repo if it doesn't exist of download it if it does.
func cloneIfNotExist() error {
	// Check if the repo already exists
	if _, err := os.Stat(getGitDir()); err == nil {
		log.Println("Repository already exists.")
		return nil
	}

	// Since it doesn't exist, we will clone it now
	log.Println("Clone repository...")
	_, err := git.PlainClone(getGitDir(), false, &git.CloneOptions{
		URL:               os.Getenv("GIT_URL"),
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return err
	}
	log.Println("Done")
	return nil
}

// Pull the repo from the origin
func pull() error {
	// Open the repo
	r, err := git.PlainOpen(getGitDir())
	if err != nil {
		return err
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

// The list is chronocally sorted with the newest commits as last
func history() ([]gitobject.Commit, error) {
	// Open the repo
	r, err := git.PlainOpen(getGitDir())
	if err != nil {
		return nil, err
	}

	// Get the HEAD
	cIter, err := r.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}

	// Make a list of commits
	var commits []gitobject.Commit
	err = cIter.ForEach(func(c *gitobject.Commit) error {
		commits = append(commits, *c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort the commits
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Author.When.Local().Unix() < commits[j].Author.When.Local().Unix()
	})

	return commits, nil
}

// Get all commits between two commits
// Since is not included, until however is
func historyBetween(since string, until string) ([]gitobject.Commit, error) {
	all, err := history()
	if err != nil {
		return nil, err
	}

	// Filter the commits
	var between []gitobject.Commit
	sawSince := false
	for _, c := range all {
		if sawSince {
			between = append(between, c)
		}

		if c.Hash.String() == since {
			sawSince = true
		}

		if c.Hash.String() == until {
			break
		}

	}

	return between, nil
}

// Get all commits since a past commit (given with since which is a commit hash)
func historySince(since string) ([]gitobject.Commit, error) {
	curr, err := getCurrentCommit()
	if err != nil {
		return nil, err
	}
	return historyBetween(since, curr)
}

// Get the username from the git URL (your username).
func getGitUser() string {
	url, _ := url.Parse(os.Getenv("GIT_URL"))
	return url.User.Username()
}

func getGitDir() string {
	return filepath.Join("data", getGitUser())
}

// Get the current commit hash
func getCurrentCommit() (string, error) {
	r, err := git.PlainOpen(getGitDir())
	if err != nil {
		return "", err
	}

	ref, err := r.Head()
	return ref.Hash().String(), nil
}

func readFile(path string) ([]byte, error) {
	// TODO avoid path tranversal attack
	return ioutil.ReadFile(filepath.Join(getGitDir(), path))
}

func listFiles(path string) ([]string, error) {
	// TODO avoid path tranversal attack
	files, err := ioutil.ReadDir(filepath.Join(getGitDir(), path))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	names := make([]string, 0, len(files))
	for _, file := range files {
		name := ""
		if file.IsDir() {
			name = "📁 "
		} else {
			name = "📄 "
		}

		name += file.Name()
		names = append(names, name)
	}

	return names, nil
}
