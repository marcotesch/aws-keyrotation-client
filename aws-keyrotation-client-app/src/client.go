package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"

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
	session, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		Config: aws.Config{
			Region: region,
		},
	})

	log.Println("Searching for specified profile")

	sectionIndex := findProfileSection(path, profile)

	log.Println("Obtaining new credentials for profile")
	credential, err := obtainCredentials(path, profile, session)

	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Writing new credentials to file")
	err = writeCredentialsFile(credential, path, sectionIndex)

	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Deleting old credentials in AWS")
	err = deleteOldCredentials(credential, session)

	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Rotated credentials successfully")
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
			break
		}

		lineNumber++
	}

	return lineNumber
}

func obtainCredentials(path string, profile string, ses *session.Session) (map[string]string, error) {
	iamClient := iam.New(ses)
	stsClient := sts.New(ses)

	stsCallerIdentityResponse, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Detailed error handling
			log.Fatal(awsErr.Error())
		}
		// Not an AWS SDK Error
		return nil, err
	}

	userName := getUserName(*stsCallerIdentityResponse.Arn)

	iamCreateAccessKeyResponse, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})

	if err != nil {
		return nil, err
	}

	cred := credentials.NewSharedCredentials(path, profile)
	credValue, err := cred.Get()

	if err != nil {
		return nil, err
	}

	credential := map[string]string{
		"AccessKeyId":     *iamCreateAccessKeyResponse.AccessKey.AccessKeyId,
		"SecretAccessKey": *iamCreateAccessKeyResponse.AccessKey.SecretAccessKey,
		"UserName":        userName,
		"OldAccessKeyId":  credValue.AccessKeyID,
	}

	return credential, nil
}

func writeCredentialsFile(credential map[string]string, path string, sectionIndex int) error {
	input, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	lines[sectionIndex] = "aws_access_key_id = " + credential["AccessKeyId"]
	lines[sectionIndex+1] = "aws_secret_access_key = " + credential["SecretAccessKey"]

	output := strings.Join(lines, "\n")

	err = ioutil.WriteFile(path, []byte(output), 0600)

	if err != nil {
		return err
	}

	return nil
}

func deleteOldCredentials(credential map[string]string, ses *session.Session) error {
	iamClient := iam.New(ses)

	_, err := iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(credential["OldAccessKeyId"]),
		UserName:    aws.String(credential["UserName"]),
	})

	if err != nil {
		return err
	}

	return nil
}

func getUserName(userArn string) string {
	userArnSplit := strings.Split(userArn, "/")
	userName := userArnSplit[len(userArnSplit)-1]

	return userName
}
