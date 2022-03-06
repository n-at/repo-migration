package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////

type applicationConfiguration struct {
	GithubToken       string
	GiteaToken        string
	GiteaOrganization string
	GiteaUrl          string
	MirrorInterval    string
	Private           bool
	Issues            bool
	Labels            bool
	Lfs               bool
	Milestones        bool
	PullRequests      bool
	Releases          bool
	Wiki              bool
}

type giteaImportRequest struct {
	AuthToken    string `json:"auth_token"`
	CloneAddress string `json:"clone_addr"`
	Service      string `json:"service"`

	Mirror         bool   `json:"mirror"`
	MirrorInterval string `json:"mirror_interval"`

	Private   bool   `json:"private"`
	RepoName  string `json:"repo_name"`
	RepoOwner string `json:"repo_owner"`

	Issues       bool `json:"issues"`
	Labels       bool `json:"labels"`
	Lfs          bool `json:"lfs"`
	Milestones   bool `json:"milestones"`
	PullRequests bool `json:"pull_requests"`
	Releases     bool `json:"releases"`
	Wiki         bool `json:"wiki"`
}

///////////////////////////////////////////////////////////////////////////////

var (
	config applicationConfiguration
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config file: %s", err)
	}
	config = applicationConfiguration{
		MirrorInterval: "24h0m0s",
		Private:        false,
		Issues:         false,
		Labels:         false,
		Lfs:            false,
		Milestones:     false,
		PullRequests:   false,
		Releases:       false,
		Wiki:           false,
	}
	if err := viper.UnmarshalKey("app", &config); err != nil {
		log.Fatalf("unable to read configration: %s", err)
	}
	if len(config.GiteaToken) == 0 {
		log.Fatalf("empty gitea token")
	}
	if len(config.GiteaUrl) == 0 {
		log.Fatalf("empty gitea url")
	}
	if len(config.GiteaOrganization) == 0 {
		log.Fatalf("empty gitea organization")
	}
}

func main() {
	fileBytes, err := os.ReadFile("repos.txt")
	if err != nil {
		log.Fatalf("unable to read repos file: %s", err)
	}
	repos := strings.Split(string(fileBytes), "\n")

	for _, repo := range repos {
		request, err := createImportRequest(repo)
		if err != nil {
			log.Errorf("unable to create import request for %s: %s", repo, err)
			continue
		}
		if err := sendRequest(request); err != nil {
			log.Errorf("unable to send import request for %s: %s", repo, err)
		}
		log.Infof("%s imported", repo)
	}
}

func createImportRequest(url string) (*giteaImportRequest, error) {
	if len(url) == 0 {
		return nil, errors.New("empty repo url")
	}

	parts := strings.Split(url, "/")
	if len(parts) != 5 {
		return nil, errors.New("malformed github url")
	}
	if parts[2] != "github.com" {
		return nil, errors.New("nt a github url")
	}
	if len(parts[3]) == 0 {
		return nil, errors.New("empty owner name")
	}
	if len(parts[4]) == 0 {
		return nil, errors.New("empty repo name")
	}

	request := &giteaImportRequest{
		AuthToken:      config.GithubToken,
		CloneAddress:   url,
		Service:        "github",
		Mirror:         true,
		MirrorInterval: config.MirrorInterval,
		Private:        config.Private,
		RepoName:       fmt.Sprintf("%s__%s", parts[3], parts[4]),
		RepoOwner:      config.GiteaOrganization,
		Issues:         config.Issues,
		Labels:         config.Labels,
		Lfs:            config.Lfs,
		Milestones:     config.Milestones,
		PullRequests:   config.PullRequests,
		Releases:       config.Releases,
		Wiki:           config.Wiki,
	}

	return request, nil
}

func sendRequest(importRequest *giteaImportRequest) error {
	url := fmt.Sprintf("%s/api/v1/repos/migrate", config.GiteaUrl)

	body, err := json.Marshal(importRequest)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("token %s", config.GiteaToken))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusCreated {
		return errors.New("wrong status " + response.Status)
	}

	return nil
}
