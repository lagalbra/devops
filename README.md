Azure DevOps
============

Like any project I have big plans that I am sure I will abandon just after writing this README. For now this repo is to extract pull requests from Azure DevOps and measure how the team is doing with respect to reviewing pull requests.

Build
-----

Clone the repo and for the first time
```bash
git clone https://github.com/abhinababasu/devops
go get -v github.com/benmatselby/go-azuredevops/azuredevops
go get -v github.com/llgcode/draw2d
go get -v github.com/Azure/azure-storage-blob-go/azblob
```
 
 Subsequently run the following

```bash
go build
````

Running
-------
You need to export the following environment variables. 

```bash
export AZUREDEVOPS_ACCOUNT=<account e.g. msazure>
export AZUREDEVOPS_PROJECT=<Project e.g. One>
export AZUREDEVOPS_TOKEN=<your Azure DevOps token>
export AZUREDEVOPS_REPO=<My cool repo>
```

See the command line help
```bash
./devops -h
```

To just see PR reviewer details for the last 100 prs
```bash
./devops -pr 100
```

Just see work item details (under-development, not yet working)
```bash
./devops -wit
```

Output
------
This fetches the last N pull requests and prints the reviwers and the number of PRs they have reviewed along. Sample output

```
Using Account=msazure, Project=One
Processing 400 completed PRs.........
PRs from 2019-04-01 00:00:00.42 +0000 UTC

REVIEWER STATS

              Trillian Astra 225 (56.2%) [############################################]
                Ford Prefect 140 (35.0%) [###############################-------------]
                 Arthur Dent 134 (33.5%) [#############################---------------]
              Slartibartfast 125 (31.2%) [###########################-----------------]
           Zaphod Beeblebrox 107 (26.8%) [####################------------------------]
```
