package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/aws/session"
)

var version = "Keyrotate Client v0.1.0b"

func main() {
	profile := flag.String("profile", "default", "Name of the AWS Profile, for which Credentials should be rotated")
	region := flag.String("region", "eu-central-1", "AWS Region identifier, for use of the specific API.")

	home := retriveHomeDir()
	credFilePathDefault := home + "/.aws/credentials"

	credFilePath := flag.String("credential-file", credFilePathDefault, "Path to your AWS Credentials File")

	showVersion := flag.Bool("version", false, "Show Keyrotate Client Version")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	rotateCredentials(*credFilePath, *profile, region)
}

func retriveHomeDir() string {
	cUser, err := user.Current()

	if err != nil {
		log.Panicln(err)
	}

	return cUser.HomeDir
}

func rotateCredentials(path string, profile string, region *string) {
	profileSectionStart := findProfileSection(path, profile)
	log.Println(profileSectionStart)

	// input, err := ioutil.ReadFile(path)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// lines := strings.Split(string(input), "\n")

	obtainCredentials(profile, region)
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

	return lineNumber
}

func obtainCredentials(profile string, region *string) (map[string]string, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		Config: aws.Config{
			Region: region,
		},
	})

	if err != nil {
		return nil, err
	}

	iamClient := iam.New(session)
	stsClient := sts.New(session)

	stsCallerIdentityResponse, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	userName := getUserName(*stsCallerIdentityResponse.Arn)

	if err != nil {
		return nil, err
	}

	iamCreateAccessKeyResponse, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})

	if err != nil {
		return nil, err
	}

	accessKey := map[string]string{
		"AccessKeyId":     *iamCreateAccessKeyResponse.AccessKey.AccessKeyId,
		"SecretAccessKey": *iamCreateAccessKeyResponse.AccessKey.SecretAccessKey,
	}

	return accessKey, nil
}

func getUserName(userArn string) string {
	userArnSplit := strings.Split(userArn, "/")
	userName := userArnSplit[len(userArnSplit)-1]

	return userName
}
