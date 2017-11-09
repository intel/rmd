package main

// An tool to help clone vendors for RMD,

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Dep of vendors
type Dep struct {
	ImportPath string
	Comment    string
	Rev        string
}

// GodepJson is struct of the whole json
type GodepJson struct {
	ImportPath   string
	GoVersion    string
	GodepVersion string
	Packages     []string
	Deps         []Dep
}

func clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	err := cmd.Run()
	return err
}

// run git command out side of git directory
func localGit(repo string, cmds ...string) error {
	gitdir := "--git-dir=" + repo + "/.git"
	gitworktree := "--work-tree=" + repo
	var args []string
	args = append(args, gitdir)
	args = append(args, gitworktree)
	for _, cmd := range cmds {
		args = append(args, cmd)
	}
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	return err
}

// Return remote repo path and local repo path
func getRepoPath(importPath string) (string, string) {
	paths := strings.Split(importPath, "/")

	// repo for golang.org/x/text was hold in go.googlesource.com
	if paths[0] == "golang.org" {
		paths = paths[:3]
		return "go.googlesource.com/" + paths[2], strings.Join(paths, "/")
	}

	// for package in google.golang.org and mgo.v2, yam.v2, special deal
	if paths[0] == "google.golang.org" {
		paths = paths[:2]
		return "github.com/golang/" + paths[1], strings.Join(paths, "/")
	}

	if paths[1] == "mgo.v2" || paths[1] == "yaml.v2" {
		paths = paths[:2]
	} else {
		// Only return reposerver/user/repo
		paths = paths[:3]
	}
	return strings.Join(paths, "/"), strings.Join(paths, "/")

}

func main() {
	var jsonPath, vendorPath string
	flag.StringVar(&jsonPath, "godeps", "./Godeps/Godeps.json", "Godep json file path")
	flag.StringVar(&vendorPath, "vendor", "vendor/", "vendor repo path")
	flag.Parse()

	if !strings.HasSuffix(vendorPath, "/") {
		vendorPath = vendorPath + "/"
	}

	raw, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		fmt.Println("error when read ", jsonPath)
		os.Exit(1)
	}

	var godepjson GodepJson

	err = json.Unmarshal(raw, &godepjson)
	if err != nil {
		fmt.Println("error when parse json ", err)
		os.Exit(1)
	}

	total := len(godepjson.Deps)
	cur := 1

	fmt.Printf("Download packages in %s\n", jsonPath)
	for _, repo := range godepjson.Deps {
		repoPath, localPath := getRepoPath(repo.ImportPath)
		gitSrc := "https://" + repoPath
		gitPath := vendorPath + localPath

		// clone from local or git server
		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			// copy from GOPATH
			if _, err = os.Stat(build.Default.GOPATH + "/src/" + localPath); !os.IsNotExist(err) {
				gitSrc = build.Default.GOPATH + "/src/" + localPath
			}
			fmt.Printf("\r%8s %70s to %-50s (%-3d/%-3d)", "cloning", gitSrc, gitPath, cur, total)
			err = clone(gitSrc, gitPath)
			if err != nil {
				fmt.Println("EEEEEEError when do clone ", repo.ImportPath, err)
				//
				break
			}
		}

		fmt.Printf("\r%8s %70s to %-50s (%-3d/%-3d)", "checkout", repo.ImportPath, repo.Rev, cur, total)
		err = localGit(gitPath, "checkout", repo.Rev)
		if err != nil {
			if err = localGit(gitPath, "fetch", "-v"); err != nil {
				fmt.Println("ERRRRRRORR to fetch origin ", repo.ImportPath, err)
			}
			// checkout again after fetch
			if err = localGit(gitPath, "checkout", repo.Rev); err != nil {
				fmt.Println("ERRRRRRORR to check", repo.ImportPath, err)
				os.Exit(1)
			}
		}
		cur++
	}

	fmt.Println()
}
