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

type WitQueryResult struct {
	QueryType string `json:"queryType"`
	AsOf      string `json:"asOf"`
	WorkItems []Wits `json:"workItems"`
}

type Wits struct {
	Id    int `json:"id"`
	State string
}

type WorkItem struct {
	Id        int    `json:"id"`
	WitFields Fields `json:"fields"`
}

type Fields struct {
	State string `json:"System.State"`
}

type WitStateCount struct {
	State string
	Count int
}

func NewWork(account, project, token, queryId string) (r *AzureDevopsWit) {
	r = &AzureDevopsWit{}
	r.client = constructClientFromConfig(account, project, token)

	URL := fmt.Sprintf(
		"_apis/wit/queries/%s?api-version=4.1",
		url.PathEscape(queryId),
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

func (r *AzureDevopsWit) RefreshWit(queryId string) ([]WitStateCount, error) {
	wits, err := r.loadWorkitems(queryId)
	if err != nil {
		return nil, err
	}

	m := make(map[string]int)
	for _, w := range wits {
		m[w.State]++
		//fmt.Println(w)
	}
	states := make([]WitStateCount, len(m))
	i := 0
	for k, v := range m {
		states[i] = WitStateCount{k, v}
		i++
	}
	return states, nil
}

func (r *AzureDevopsWit) loadWorkitems(queryId string) ([]Wits, error) {
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/wit/wiql/query%20by%20id?view=azure-devops-rest-4.1
	URL := fmt.Sprintf("_apis/wit/wiql/%s?api-version=4.1", queryId)

	request, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var response WitQueryResult // PullRequestsResponse
	_, err = r.client.Execute(request, &response)
	if err != nil {
		return nil, err
	}

	for i, w := range response.WorkItems {
		wi, err := r.getWorkitem(w.Id)
		if err != nil {
			fmt.Printf("Error fetching workitem %v: %v", wi.Id, err)
			w.State = "Unknown"
		} else {
			response.WorkItems[i].State = wi.WitFields.State
		}
	}
	return response.WorkItems, nil
}

func (r *AzureDevopsWit) getWorkitem(witId int) (WorkItem, error) {
	var wi WorkItem
	URL := fmt.Sprintf("_apis/wit/workitems/%v?api-version=4.1", witId)

	req, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Println(err)
		return wi, err
	}

	_, err = r.client.Execute(req, &wi)
	if err != nil {
		return wi, err
	}

	return wi, nil
}
