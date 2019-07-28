package main

// Docs
// https://docs.microsoft.com/en-us/rest/api/azure/devops/wit/?view=azure-devops-rest-4.1
import (
	"fmt"
	"net/url"

	az "github.com/benmatselby/go-azuredevops/azuredevops"
)

type AzureDevopsWit struct {
	client *az.Client
	Query  WitQuery
	err    error
}

type WitQuery struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func NewWork(account, project, token, queryName string) (r *AzureDevopsWit) {
	r = &AzureDevopsWit{}
	r.client = constructClientFromConfig(account, project, token)

	URL := fmt.Sprintf(
		"_apis/wit/queries/%s?api-version=4.1",
		url.PathEscape(queryName),
	)

	var witQuery WitQuery
	request, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		r.err = err
		return
	}

	_, err = r.client.Execute(request, &witQuery)

	if err != nil {
		r.err = err
	}
	r.Query = witQuery

	return
}

func (r *AzureDevopsRepo) RefreshWit(count int) {
	var errs []error
	if err := r.loadPullRequests(count); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		r.err = fmt.Errorf("Error(s) occurred: %v", errs)
	}

	return
}
