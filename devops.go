package main

import (
	"fmt"
	"os"
)

// Docs
// https://docs.microsoft.com/en-us/rest/api/azure/devops/git/pull%20requests/get%20pull%20requests?view=azure-devops-server-rest-4.1
func main() {
	// Fetch the access stuff from environment
	acc := os.Getenv("AZUREDEVOPS_ACCOUNT")
	proj := os.Getenv("AZUREDEVOPS_PROJECT")
	token := os.Getenv("AZUREDEVOPS_TOKEN")
	repo := os.Getenv("AZUREDEVOPS_REPO")

	// Connect to repo
	r := NewRepo(acc, proj, token, repo)
	if r.err != nil {
		fmt.Println(r.err)
		return
	}

	// Fetch PRs
	count := 400
	prStats, max := r.GetPullRequestReviewsByUser(count)
	barmax := float32(100.0)
	// Output!!
	for _, ps := range prStats {
		bar := int((barmax / float32(max)) * float32(ps.Count))
		fmt.Printf("%30s %4d ", ps.Name, ps.Count)
		for i := 0; i < bar; i++ {
			fmt.Print("*")
		}

		fmt.Println()
	}

}
