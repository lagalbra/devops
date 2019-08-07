package main

// Docs
// https://docs.microsoft.com/en-us/rest/api/azure/devops/wit/?view=azure-devops-rest-4.1
import (
	"fmt"
	"time"

	az "github.com/benmatselby/go-azuredevops/azuredevops"
)

type AzureDevopsWit struct {
	client *az.Client
}

type WitQuery struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type WitQueryResult struct {
	QueryType    string              `json:"queryType"`
	AsOf         string              `json:"asOf"`
	WorkItems    []WorkItem          `json:"workItems,omitempty"`
	WitRelations []WorkItemRelations `json:"workItemRelations,omitempty"`
}

type WorkItemRelations struct {
	Target WitTarget `json:"target"`
}

type WitTarget struct {
	Id int `json:"id"`
}

type WorkItem struct {
	Id          int `json:"id"`
	State       string
	Type        string
	Title       string
	AssignedTo  string
	ChangedDate time.Time
}

type WorkItemInternal struct {
	Id        int    `json:"id"`
	WitFields Fields `json:"fields"`
}

type Fields struct {
	State       string `json:"System.State"`
	Type        string `json:"System.WorkItemType"`
	Title       string `json:"System.Title"`
	AssignedTo  string `json:"System.AssignedTo"`
	ChangedDate string `json:"System.ChangedDate"`
}

type WiqlQuery struct {
	Query string `json:"query"`
}

type EpicStat struct {
	Epic       WorkItem
	Done       int
	NotDone    int
	InProgress int
	Unknown    int
}

func NewWork(account, project, token string) (r *AzureDevopsWit) {
	r = &AzureDevopsWit{}
	r.client = constructClientFromConfig(account, project, token)

	return
}

func (r *AzureDevopsWit) GetWorkitems(queryId string) ([]WorkItem, error) {
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/wit/wiql/query%20by%20id?view=azure-devops-rest-4.1
	URL := fmt.Sprintf("_apis/wit/wiql/%s?api-version=4.1", queryId)

	request, err := r.client.NewRequest("GET", URL, nil)

	if err != nil {
		return nil, err
	}

	var response WitQueryResult // PullRequestsResponse
	_, err = r.client.Execute(request, &response)
	if err != nil {
		return nil, err
	}

	var workItems []WorkItem

	for _, w := range response.WorkItems {
		wi, err := r.GetWorkitem(w.Id)
		if err != nil {
			return nil, err
		}
		workItems = append(workItems, wi)
	}

	return workItems, nil
}

func (q *AzureDevopsWit) RefreshWit(parentEpic int, filterSemester bool) (EpicStat, error) {

	epic, err := q.GetWorkitem(parentEpic)
	if err != nil {
		return EpicStat{}, err
	}

	wits, err := q.loadWorkitems(parentEpic)
	epicStat := EpicStat{epic, 0, 0, 0, 0}
	if err != nil {
		return epicStat, err
	}

	now := time.Now()
	month := now.Month()
	if month < 7 {
		month = time.January // First semester
	} else {
		month = time.July // Second semester
	}

	// Starttime of the current semester
	semesterStart := time.Date(now.Year(), month, 1, 0, 0, 0, 0, time.Local)

	for _, w := range wits {
		if w.Type == "Epic" { // don't count the epics
			continue
		}

		if filterSemester && (w.State == "Done" || w.State == "Removed") &&
			w.ChangedDate.Before(semesterStart) {
			continue
		}

		switch w.State {
		case "New", "To Do", "Committed":
			epicStat.NotDone++
		case "In Progress":
			epicStat.InProgress++
		case "Done", "Removed":
			epicStat.Done++
		default:
			epicStat.Unknown++
		}

	}
	return epicStat, nil
}

func (r *AzureDevopsWit) loadWorkitems(parentEpic int) ([]WorkItem, error) {
	// https://docs.microsoft.com/en-us/rest/api/azure/devops/wit/wiql/query%20by%20id?view=azure-devops-rest-4.1
	URL := "_apis/wit/wiql?api-version=4.1"

	body := `
	SELECT
    [System.Id],
    [System.WorkItemType],
    [System.Title],
    [System.AssignedTo],
    [System.State],
	[System.Tags],
	[System.ChangedDate]
	FROM workitemLinks
	WHERE
		(
			[Source].[System.TeamProject] = @project
			AND [Source].[System.WorkItemType] <> ''
			AND [Source].[System.Id] = %v
		)
		AND (
			[System.Links.LinkType] = 'System.LinkTypes.Hierarchy-Forward'
		)
		AND (
			[Target].[System.TeamProject] = @project
			AND [Target].[System.WorkItemType] <> ''
			AND NOT [Target].[System.State] IN ('Removed')
			AND [Target].[Scrum_custom.Status] <> 'Deferred'
		)
	MODE (Recursive)
	`
	var wiqlQuery WiqlQuery
	wiqlQuery.Query = fmt.Sprintf(body, parentEpic)
	request, err := r.client.NewRequest("POST", URL, wiqlQuery)

	if err != nil {
		return nil, err
	}

	var response WitQueryResult // PullRequestsResponse
	_, err = r.client.Execute(request, &response)
	if err != nil {
		return nil, err
	}

	var workItems []WorkItem

	for _, w := range response.WitRelations {
		wi, err := r.GetWorkitem(w.Target.Id)
		if err != nil {
			return nil, err
		}

		workItems = append(workItems, wi)
	}

	return workItems, nil
}

func (r *AzureDevopsWit) GetWorkitem(witId int) (WorkItem, error) {
	var wi WorkItemInternal
	URL := fmt.Sprintf("_apis/wit/workitems/%v?api-version=4.1", witId)

	req, err := r.client.NewRequest("GET", URL, nil)
	if err != nil {
		return WorkItem{}, err
	}

	_, err = r.client.Execute(req, &wi)
	if err != nil {
		return WorkItem{}, err
	}

	t, _ := time.Parse(time.RFC3339, wi.WitFields.ChangedDate)
	return WorkItem{wi.Id, wi.WitFields.State, wi.WitFields.Type, wi.WitFields.Title, wi.WitFields.AssignedTo, t}, nil
}
