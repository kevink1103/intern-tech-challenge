package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version
	// This is just an example structure of the code, if you implement this interface, the test cases in main_test.go are very easy to run

	// Sort in ascending order
	semver.Sort(releases)

	// Access each element in reverse order
	// To make it descending order and
	// Append versionSlice at the same time
	// Only one loop needed O(n)
	for index := range releases {
		var release = releases[len(releases)-index-1]
		// Filtering pre-releases and lower versions
		if release.PreRelease != "" || release.LessThan(*minVersion) {
			continue
		}

		if len(versionSlice) == 0 || versionSlice[len(versionSlice)-1].Major != release.Major || versionSlice[len(versionSlice)-1].Minor != release.Minor {
			versionSlice = append(versionSlice, release)
		}
	}

	return versionSlice
}

// CheckGithubRepo gets list of releases and prints last versions
func CheckGithubRepo(ctx context.Context, client *github.Client, user string, repo string, minVer string) {
	opt := &github.ListOptions{PerPage: 50}
	releases, _, err := client.Repositories.ListReleases(ctx, user, repo, opt)
	if err != nil {
		// panic(err) // is this really a good way?
		fmt.Printf("%s\r\n", err.Error())
		return
	}
	minVersion := semver.New(minVer)
	allReleases := make([]*semver.Version, len(releases))
	for i, release := range releases {
		versionString := *release.TagName
		if versionString[0] == 'v' {
			versionString = versionString[1:]
		}
		allReleases[i] = semver.New(versionString)
	}

	versionSlice := LatestVersions(allReleases, minVersion)

	fmt.Printf("latest versions of %s/%s: %s\r\n", user, repo, versionSlice)
}

// ReadFile reads the file and returns slices [[repositories, min_version]]
func ReadFile(fileName string) [][]string {
	var fileData [][]string

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	for scanner.Scan() {
		singleLine := scanner.Text()
		if strings.Contains(singleLine, "repository,min_version") {
			continue
		}

		tokens := strings.Split(singleLine, ",")
		names := strings.Split(tokens[0], "/")

		if len(tokens) == 2 && len(names) == 2 {
			fileData = append(fileData, []string{names[0], names[1], tokens[1]})
		} else {
			fmt.Printf("rewrite this line: %s\r\n", singleLine)
		}
	}

	return fileData
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	fileName := os.Args[1]
	fileData := ReadFile(fileName)

	// Github
	// Create API client once here for optimization
	client := github.NewClient(nil)
	ctx := context.Background()
	for _, repo := range fileData {
		CheckGithubRepo(ctx, client, repo[0], repo[1], repo[2])
	}
}
