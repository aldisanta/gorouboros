package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	branch := os.Args[1]
	branches := strings.Split(branch, "/")
	id := branches[1]

	var (
		cmdOut       []byte
		err          error
		cmd          strings.Builder
		prevCommitID string
		nextCommitID string
		commitMsg    string
	)
	// get commit ID
	cmd.WriteString("git log | head -n 1 | cut -c 8-")
	if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
		os.Exit(1)
	}
	commitID := string(cmdOut)
	cmd.Reset()

	// get latest commit has
	cmd.WriteString("git rev-parse HEAD")
	if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
		os.Exit(1)
	}
	lastCommitHash := strings.TrimSpace(string(cmdOut))
	cmd.Reset()

	// get commit works, from latest commit hash
	cmd.WriteString("git log")
	cmd.WriteString(" ")
	cmd.WriteString(branch)
	cmd.WriteString(" ")
	cmd.WriteString(lastCommitHash)
	cmd.WriteString("..HEAD")
	cmd.WriteString(" ")
	cmd.WriteString("--grep=\"\\[\\")
	cmd.WriteString(id)
	cmd.WriteString("\\]\" | grep \"commit\" | cut -c 8-")
	if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
		os.Exit(1)
	}
	commitIDs := strings.Split(string(cmdOut), "\n")
	cmd.Reset()

	// get message works
	cmd.WriteString("git log")
	cmd.WriteString(" ")
	cmd.WriteString(branch)
	cmd.WriteString(" ")
	cmd.WriteString(lastCommitHash)
	cmd.WriteString("..HEAD")
	cmd.WriteString(" ")
	cmd.WriteString("| grep \"\\[\\")
	cmd.WriteString(id)
	cmd.WriteString("\\]\" | cut -c 5-")
	if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
		os.Exit(1)
	}
	commitMsgs := strings.Split(string(cmdOut), "\n")
	cmd.Reset()

	// reverse loop
	for idx := len(commitMsgs) - 1; idx >= 0; idx-- {
		if len(commitIDs[idx]) > 0 && len(commitMsgs[idx]) > 0 {
			if idx == len(commitMsgs)-1 {
				prevCommitID = commitID
			} else {
				prevCommitID = commitIDs[idx+1]
			}
			nextCommitID = commitIDs[idx]
			commitMsg = commitMsgs[idx]

			// get deleted commit files & check if it's deletion
			cmd.WriteString("git diff --stat")
			cmd.WriteString(" ")
			cmd.WriteString(prevCommitID)
			cmd.WriteString(" ")
			cmd.WriteString(nextCommitID)
			cmd.WriteString(" ")
			cmd.WriteString("--name-only --diff-filter=D")
			if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
				fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
				os.Exit(1)
			}
			cmd.Reset()

			// running delete
			if len(string(cmdOut)) > 0 {
				deleteFilesArray := strings.Split(string(cmdOut), "\n")
				for _, fileName := range deleteFilesArray {
					os.Remove(fileName)
				}
			}

			// get commit files & exclude Delete, Rename
			cmd.WriteString("git diff --stat")
			cmd.WriteString(" ")
			cmd.WriteString(prevCommitID)
			cmd.WriteString(" ")
			cmd.WriteString(nextCommitID)
			cmd.WriteString(" ")
			cmd.WriteString("--name-only --diff-filter=dr")
			if cmdOut, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
				fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
				os.Exit(1)
			}

			// split and join
			commitFiles := strings.Join(strings.Split(string(cmdOut), "\n"), " ")
			// gorouboros
			commitFiles = strings.Replace(commitFiles, "gorouboros", "", -1)
			// gitignore
			commitFiles = strings.Replace(commitFiles, ".gitignore", "", -1)
			cmd.Reset()

			if len(commitFiles) > 0 {
				// checkout from hash specify files
				cmd.WriteString("git checkout")
				cmd.WriteString(" ")
				cmd.WriteString(nextCommitID)
				cmd.WriteString(" ")
				cmd.WriteString("--")
				cmd.WriteString(" ")
				cmd.WriteString(commitFiles)
				if _, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
					fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
					os.Exit(1)
				}
				cmd.Reset()

				// commit files with messages
				cmd.WriteString("git add . && git commit -m")
				cmd.WriteString(" ")
				cmd.WriteString("\"")
				cmd.WriteString(commitMsg)
				cmd.WriteString("\"")
				if _, err = exec.Command("bash", "-c", cmd.String()).Output(); err != nil {
					fmt.Fprintln(os.Stderr, "There was an error running command: ", cmd.String(), " || error: ", err)
					os.Exit(1)
				}
				cmd.Reset()

				if idx > 0 {
					rand.Seed(time.Now().UnixNano())
					time.Sleep(time.Duration(rand.Int63n(5-3+1)+3) * time.Minute)
				}
			}
		} else {
			fmt.Println("Nothing to process")
		}
	}

}
