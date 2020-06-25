package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"time"
)

var rootDir = flag.String("rootDir", "/tmp", "Directory to scan for SSH socket files")
var regex = flag.String("regex", "ssh-.*", "Regular expression to match subdirectory names against")

func getSocketFiles(rootDir, regex *string) (error, []string) {
	var socketFiles []string

	files, err := ioutil.ReadDir(*rootDir)
	if err != nil {
		return err, []string{}
	}

	for _, file := range files {
		if file.IsDir() {
			re := regexp.MustCompile(*regex)
			if re.MatchString(file.Name()) {
				baseDir := fmt.Sprintf("%s/%s", *rootDir, file.Name())
				fs, err := ioutil.ReadDir(baseDir)
				if err != nil {
					return err, []string{}
				}
				if len(fs) == 0 {
					err = fmt.Errorf("No files found in %s", baseDir)
					return err, []string{}
				}
				if len(fs) > 1 {
					err = fmt.Errorf("More than 1 file found in %s", baseDir)
					return err, []string{}
				}
				if fs[0].Mode()&os.ModeSocket != 0 {
					path := fmt.Sprintf("%s/%s/%s", *rootDir, file.Name(), fs[0].Name())
					socketFiles = append(socketFiles, path)
				}
			}
		}
	}

	return nil, socketFiles
}

func getGoodSocketFile(socketFiles []string) (error, string) {
	goodSocketFile := ""
	for _, socketFile := range socketFiles {
		// Start ssh-add
		cmd := exec.Command("ssh-add", "-l")
		cmd.Env = []string{fmt.Sprintf("SSH_AUTH_SOCK=%s", socketFile)}
		if err := cmd.Start(); err != nil {
			return err, ""
		}

		// Wait for ssh-add to finish, or kill it after 2 seconds
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-time.After(2 * time.Second):
			if err := cmd.Process.Kill(); err != nil {
				return err, ""
			}
			fmt.Fprintf(os.Stderr, "Deleting stale socket file %s\n", socketFile)
			err := os.Remove(socketFile)
			if err != nil {
				return err, ""
			}
			err = os.Remove(path.Dir(socketFile))
			if err != nil {
				return err, ""
			}
		case err := <-done:
			if err != nil {
				return err, ""
			}
			goodSocketFile = socketFile
		}
	}

	return nil, goodSocketFile
}

func main() {
	flag.Parse()

	err, socketFiles := getSocketFiles(rootDir, regex)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch length := len(socketFiles); {
	case length < 1:
		// fmt.Fprintln(os.Stderr, "No sockets found")
		os.Exit(1)
	case length == 1:
		fmt.Printf("%s", socketFiles[0])
	case length > 1:
		err, goodSocketFile := getGoodSocketFile(socketFiles)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("%s", goodSocketFile)
	}
}
