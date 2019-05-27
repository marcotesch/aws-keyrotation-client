package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"regexp"
)

var version = "Keyrotate Client v0.1.0b"

func main() {
	profile := flag.String("profile", "default", "Name of the AWS Profile, for which Credentials should be rotated")

	home := retriveHomeDir()
	credFilePathDefault := home + "/.aws/credentials"

	credFilePath := flag.String("credential-file", credFilePathDefault, "Path to your AWS Credentials File")

	showVersion := flag.Bool("version", false, "Show Keyrotate Client Version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	rotateCredentials(*credFilePath, *profile)
}

func retriveHomeDir() string {
	cUser, err := user.Current()

	if err != nil {
		log.Panicln(err)
	}

	return cUser.HomeDir
}

func rotateCredentials(path string, profile string) {
	findProfileSection(path, profile)
}

func findProfileSection(path string, profile string) int {
	matcher, err := regexp.Compile(`(?m)^(\[` + profile + `\])$`)

	if err != nil {
		log.Panicln(err)
	}

	file, err := os.Open(path)

	if err != nil {
		log.Panicln(err)
	}

	fileScanner := bufio.NewScanner(file)
	lineNumber := 1

	for fileScanner.Scan() {

		lineText := fileScanner.Text()

		if matcher.Match([]byte(lineText)) {
			log.Printf("Profile Section found in line %d\n", lineNumber)
			break
		}

		lineNumber++
	}

	return 0
}
