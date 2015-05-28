package main

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"regexp"
)

var (
	hookbotGithostRe = regexp.MustCompile("^/sub/([^/]+)/repo/([^/]+)/([^/]+)" +
		"/branch/([^/]+)$")
	hookbotDockerPullSub = regexp.MustCompile("^/sub/docker-pull/(.*)/tag/([^/]+)$")
	hookbotDockerPullPub = regexp.MustCompile("^/pub/docker-pull/(.*)/tag/([^/]+)$")
)

func GetSourceFromHookbot(hookbotURLStr string) (string, ImageSource, error) {

	hookbotURL, err := url.Parse(hookbotURLStr)
	if err != nil {
		return "", nil, fmt.Errorf("Hookbot URL %q does not parse: %v",
			hookbotURLStr, err)
	}

	switch {
	case hookbotGithostRe.MatchString(hookbotURL.Path):
		return NewGitHostSource(hookbotURL)

	case hookbotDockerPullSub.MatchString(hookbotURL.Path):
		return NewDockerPullSource(hookbotURL)
	}

	return "", nil, fmt.Errorf("Unrecogized hookbot URL %q", hookbotURL.Path)
}

func NewGitHostSource(hookbotURL *url.URL) (string, ImageSource, error) {

	groups := hookbotGithostRe.FindStringSubmatch(hookbotURL.Path)
	host, user, repository, branch := groups[1], groups[2], groups[3], groups[4]

	imageSource := &GitHostSource{
		Host:          host,
		User:          user,
		Repository:    repository,
		InitialBranch: branch,
	}

	log.Printf("Hookbot monitoring %v@%v via %v",
		repository, branch, hookbotURL.Host)

	return repository, imageSource, nil
}

func NewDockerPullSource(hookbotURL *url.URL) (string, ImageSource, error) {

	groups := hookbotDockerPullSub.FindStringSubmatch(hookbotURL.Path)
	repository, tag := groups[1], groups[2]

	imageSource := &DockerPullSource{
		Repository: repository,
		Tag:        tag,
	}

	log.Printf("Hookbot monitoring %v@%v via %v",
		repository, tag, hookbotURL.Host)

	containerName := path.Base(repository)
	return containerName, imageSource, nil
}

func ParseHookbotDockerPullPubEndpoint(hookbotURLStr string) (image, tag string, err error) {
	u, err := url.Parse(hookbotURLStr)
	if err != nil {
		return "", "", err
	}

	parts := hookbotDockerPullPub.FindStringSubmatch(u.Path)
	if parts == nil {
		return "", "", fmt.Errorf("Pub URL %q doesn't match: %q",
			u.Path, hookbotDockerPullPub.String())
	}

	return parts[1], parts[2], nil
}