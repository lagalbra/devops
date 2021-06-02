package main

// Docs
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20requests?view=azure-devops-server-rest-4.1

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

type CommitResponse struct {
	Count   int      `json:"count"`
	Commits []Commit `json:"value"`
}

type Commit struct {
	ID     string     `json:"commitId"`
	Author AuthorInfo `json:"author"`
	// Ignored fields:
	// 	committer AuthorInfo
	// 	comment
	// 	changeCounts struct
	// 	changes array
	// 	url
}

type AuthorInfo struct {
	Name string `json:"name"`
	Date string `json:"date"`
	// email is also included, but will be ignored
}

type NameStat struct {
	Name  string
	Count int
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

func (r *AzureDevopsRepo) GetPullRequestReviewsByUser(count int) ([]NameStat, int) {
	Info.Printf("Processing %v completed PRs", count)
	r.Refresh(count)
	prs := r.PullRequests

	Info.Println("PRs from", prs[len(prs)-1].ClosedDate)

	// Iterate and create a map of reviewers[review-count]
	review := make(map[string]int)
	for _, pr := range prs {
		for _, rv := range pr.Reviewers {
			// filter for specific user and ensure we do not count PR creater approving their own PR
			if !strings.Contains(rv.DisplayName, "AzLinux SAP HANA RP Devs") && rv.Vote != 0 && rv.DisplayName != pr.CreatedBy.DisplayName {
				review[rv.DisplayName]++
			}
		}
	}

	// Sort the PRs by review count, by stuffing into a slice
	max := 0
	var reviewerStat []NameStat
	for k, v := range review {
		reviewerStat = append(reviewerStat, NameStat{k, v})
		if v > max {
			max = v
		}
	}

	sort.Slice(reviewerStat, func(i, j int) bool {
		return reviewerStat[i].Count > reviewerStat[j].Count
	})

	return reviewerStat, max
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

const commitFromDateExpectedTimeFormat = "1/2/2006 03:04:05 PM" // expected format like "6/14/2018 12:00:00 AM", see https://docs.microsoft.com/en-us/rest/api/azure/devops/git/commits/get%20commits?view=azure-devops-rest-6.0#in-a-date-range

// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/commits/get%20commits?view=azure-devops-rest-6.0#on-a-branch-and-in-a-path
func (r *AzureDevopsRepo) GetCommitsByAuthor(
	branch string,
	path string,
	fromDate time.Time) ([]Commit, error) {
	params := url.Values{}
	params.Add("searchCriteria.itemPath", path)
	params.Add("searchCriteria.itemVersion.version", branch)
	params.Add("searchCriteria.fromDate", fromDate.Format(commitFromDateExpectedTimeFormat))
	params.Add("searchCriteria.$top", "1000") // set this very high because pagination is not explained: https://docs.microsoft.com/en-us/rest/api/azure/devops/git/commits/get%20commits?view=azure-devops-rest-6.0#paging

	URL := fmt.Sprintf(
		"/_apis/git/repositories/%s/commits?%s&%s",
		r.Repo.ID,
		params.Encode(),
		"api-version=6.0",
	)

	request, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to repository: %+v", err)
	}

	var response CommitResponse
	_, err = r.client.Execute(request, &response)
	if err != nil {
		return nil, fmt.Errorf("error executing request for commits: %+v", err)
	}

	return response.Commits, nil
}

func constructClientFromConfig(account, project, token string) *az.Client {
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
