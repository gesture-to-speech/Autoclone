package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

type User struct {
	Email string
	Name  string
}

type Config struct {
	PushFolder  string
	PullFolder  string
	SshPushBase string
	Users       []User
	Repos       []struct {
		Ssh string
		Key string
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Open the file
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("Error reading file: %s", err)
		return
	}

	// Unmarshal the JSON data into the Config struct
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("Error parsing JSON: %s", err)
		return
	}

	/*
		// remove pull and push folders
		err = executeCommand("", "rm", "-rf", fmt.Sprintf("\"%s\"", config.PushFolder))
		if err != nil {
			log.Printf("Error removing push folder %s", err)
			return
		}
		err = executeCommand("", "rm", "-rf", fmt.Sprintf("\"%s\"", config.PullFolder))
		if err != nil {
			log.Printf("Error removing pull folder %s", err)
			return
		}
	*/
	// create pull and push folders
	err = executeCommand("", "mkdir", "-p", config.PushFolder)
	if err != nil {
		log.Printf("Error creating push folder %s", err)
		return
	}
	err = executeCommand("", "mkdir", "-p", config.PullFolder)
	if err != nil {
		log.Printf("Error creating pull folder %s", err)
		return
	}

	// clone repos
	for _, repo := range config.Repos {
		log.Printf("Cloning repo %s", repo.Ssh)
		err = cloneRepo(config.PullFolder, repo.Ssh)
		if err != nil {
			log.Printf("Error cloning repo %s; error message: %s", repo.Ssh, err)
			return
		}

		gitLabRepo := config.SshPushBase + getRepoName(repo.Ssh) + ".git"
		err = cloneRepo(config.PushFolder, gitLabRepo)
		if err != nil {
			log.Printf("Error cloning repo %s; error message: %s", gitLabRepo, err)
			return
		}

		err = fetchOrigin(config.PullFolder, repo.Ssh)
		if err != nil {
			log.Printf("Error fetching origin repo %s; error message: %s", repo.Ssh, err)
			return
		}
		err = fetchOrigin(config.PushFolder, gitLabRepo)
		if err != nil {
			log.Printf("Error fetching origin repo %s; error message: %s", gitLabRepo, err)
			return
		}

		log.Print("Getting branches")
		allBranches, err := getAllBranches(config.PullFolder, repo.Ssh)
		if err != nil {
			log.Printf("Couldn't get all branches from repo %s; error message: %s", repo.Ssh, err)
			return
		}

		originRepoDir := getRepoFolder(repo.Ssh, config.PullFolder)
		destRepoDir := getRepoFolder(gitLabRepo, config.PushFolder)
		for _, branch := range allBranches {
			err = setUser(config.Users, destRepoDir)
			if err != nil {
				log.Printf("Couldn't set user for repo %s; error message: %s", repo.Ssh, err)
				return
			}

			log.Printf("Copy branch %s from repo %s", branch, repo.Ssh)
			err = copyBranch(branch, originRepoDir, destRepoDir)
			if err != nil {
				log.Printf("Couldn't copy branch %s from repo %s; error message: %s", branch, repo.Ssh, err)
				return
			}
		}
	}
}

func executeCommand(dir string, commandName string, arg ...string) error {
	cmd := exec.Command(commandName, arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func getRepoName(ssh string) string {
	repoSSHParts := strings.Split(ssh, "/")
	repoName := repoSSHParts[len(repoSSHParts)-1]
	return strings.TrimSuffix(repoName, ".git")
}

func getRepoFolder(ssh string, folder string) string {
	repoName := getRepoName(ssh)

	return folder + repoName + "/"
}

func setUser(users []User, dir string) error {
	numUsers := len(users)
	userId := rand.Intn(numUsers)
	log.Printf("Chosen user: %s; %s", users[userId].Name, users[userId].Email)
	err := executeCommand(dir, "git", "config", "user.email", users[userId].Email)
	if err != nil {
		return err
	}

	return executeCommand(dir, "git", "config", "user.name", users[userId].Name)
}

func copyFiles(destDir string, originDir string) error {
	// remove old files
	_, err := os.Stat(destDir)
	if !os.IsNotExist(err) {
		err = executeCommand(
			"",
			"find",
			destDir,
			"-mindepth",
			"1",
			"-not",
			"-path",
			destDir+".git/*",
			"-not",
			"-path",
			destDir+".git",
			"-delete",
		)
		if err != nil {
			return err
		}
	}

	// copy files
	err = executeCommand(
		"",
		"rsync",
		"-av",
		"--exclude=.git",
		originDir,
		destDir,
	)

	return err
}

func copyBranch(branch string, originDir string, destDir string) error {
	log.Print("Pull changes and set correct branch")
	err := executeCommand(originDir, "git", "checkout", branch)
	if err != nil {
		return err
	}

	err = executeCommand(originDir, "git", "pull", "origin", branch)
	if err != nil {
		return err
	}

	err = executeCommand(destDir, "git", "checkout", branch)
	if err != nil {
		err = executeCommand(destDir, "git", "checkout", "-b", branch)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Pull updates for %s", destDir)
		err = executeCommand(destDir, "git", "pull", "origin", branch)
		if err != nil {
			return err
		}
	}

	log.Print("Copy files")
	err = copyFiles(destDir, originDir)
	if err != nil {
		return err
	}

	log.Print("Pushing changes to repository")
	err = executeCommand(destDir, "git", "add", ".")
	if err != nil {
		log.Print("No changes")
		return nil
	}

	err = executeCommand(destDir, "git", "commit", "-m", "Update")
	if err != nil {
		log.Print("No changes")
		return nil
	}

	return executeCommand(destDir, "git", "push", "origin", branch)
}

func cloneRepo(dir string, repoSSH string) error {
	repoFolder := getRepoFolder(repoSSH, dir)

	_, err := os.Stat(repoFolder)
	if !os.IsNotExist(err) {
		log.Print("Repository already initialized")
		return nil
	}

	log.Printf("Initializing repository %s", repoSSH)
	err = executeCommand(dir, "git", "clone", repoSSH)
	return err
}

func getAllBranches(pullDir string, repoSSH string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-a")
	cmd.Dir = getRepoFolder(repoSSH, pullDir)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var branches []string
	for _, line := range lines {
		if strings.HasPrefix(line, "*") {
			line = strings.TrimPrefix(line, "*")
		}
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, " ") && strings.Contains(line, "remotes/origin/") {
			line = strings.TrimPrefix(line, "remotes/origin/")
			branches = append(branches, line)
		}
	}

	return branches, nil
}

func fetchOrigin(dir string, repoSSH string) error {
	return executeCommand(getRepoFolder(repoSSH, dir), "git", "fetch", "--all")
}
