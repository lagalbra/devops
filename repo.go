package main

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	az "github.com/benmatselby/go-azuredevops/azuredevops"
)

type AzureDevopsRepo struct {
	client       *az.Client
	Repo         Repository
	PullRequests []PullRequest

	// if an operation resulted in an error, it should be stored here
	// so that it can be displayed
	err error
}

type PullRequestsResponse struct {
	PullRequests []PullRequest `json:"value"`
	Count        int           `json:"count"`
}

type PullRequest struct {
	ID          int                `json:"pullRequestId,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	Created     string             `json:"creationDate"`
	CreatedBy   User               `json:"createdBy"`
	ClosedDate  time.Time          `json:closedDate`
	Repo        az.PullRequestRepo `json:"repository"`
	URL         string             `json:"url"`
	RemoteURL   string             `json:"remoteUrl"`
	Reviewers   []User             `json:"reviewers"`
}

// Repository represents a repository used by a build definition
type Repository struct {
	ID                 string                 `json:"id,omitempty"`
	Type               string                 `json:"type,omitempty"`
	Name               string                 `json:"name,omitempty"`
	URL                string                 `json:"url,omitempty"`
	RootFolder         string                 `json:"root_folder"`
	Properties         map[string]interface{} `json:"properties"`
	Clean              string                 `json:"clean"`
	DefaultBranch      string                 `json:"default_branch"`
	CheckoutSubmodules bool                   `json:"checkout_submodules"`
	RemoteUrl          string                 `json:"remoteUrl"`
}

type User struct {
	Vote          int    `json:"vote,omitempty"`
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	UniqueName    string `json:"uniqueName"`
	IsAadIdentity bool   `json:"isAadIdentity"`
	IsContainer   bool   `json:"isContainer"`
	ImageUrl      string `json:imageUrl`
}

func NewRepo(account, project, token, repoName string) (r *AzureDevopsRepo) {
	r = &AzureDevopsRepo{}
	r.client = constructClientFromConfig(account, project, token)

	URL := fmt.Sprintf(
		"_apis/git/repositories/%s?api-version=4.1",
		url.PathEscape(repoName),
	)

	var azrepo Repository
	request, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		r.err = err
		return
	}

	_, err = r.client.Execute(request, &azrepo)
	if err != nil {
		r.err = err
	}
	r.Repo = azrepo

	return
}

func (r *AzureDevopsRepo) Refresh(count int) {
	var errs []error
	if err := r.loadPullRequests(count); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		r.err = fmt.Errorf("Error(s) occurred: %v", errs)
	}

	return
}

func (r *AzureDevopsRepo) loadPullRequests(count int) error {
	params := url.Values{}
	params.Add("searchCriteria.repositoryId", r.Repo.ID)
	params.Add("searchCriteria.status", "completed")
	params.Add("$top", strconv.Itoa(count))

	URL := fmt.Sprintf(
		"/_apis/git/pullrequests?%s&%s",
		"api-version=4.1",
		params.Encode(),
	)

	request, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		return err
	}

	var response PullRequestsResponse
	_, err = r.client.Execute(request, &response)
	if err != nil {
		return err
	}

	r.PullRequests = response.PullRequests
	return nil
}

func constructClientFromConfig(account, project, token string) *az.Client {
	fmt.Printf("Using Account=%v, Project=%v\n", account, project)
	return az.NewClient(account, project, token)
}

func containsUser(name string, Users ...User) bool {
	for _, user := range Users {
		if user.DisplayName == name || user.UniqueName == name {
			return true
		}
	}
	return false
}
