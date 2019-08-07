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
build.sh
````

To also publish the images use
```bash
build.sh -p
```

Running Server
--------------
You need to export the following environment variables. 

```bash
export AZUREDEVOPS_ACCOUNT=<account e.g. msazure>
export AZUREDEVOPS_PROJECT=<Project e.g. One>
export AZUREDEVOPS_TOKEN=<your Azure DevOps token>
export AZUREDEVOPS_REPO=<My cool repo>

export AZURE_STORAGE_ACCOUNT="<your account>"
export AZURE_STORAGE_ACCESS_KEY="<key>"
```

See the command line help
```bash
./devops -h
```

To start the server in verbose mode on port 8080
```bash
./devops -v -port 8080
```

To run using the docker container
```bash
docker run -it --rm -e AZUREDEVOPS_ACCOUNT -e AZUREDEVOPS_PROJECT \
      -e AZUREDEVOPS_TOKEN -e AZUREDEVOPS_REPO -e AZURE_STORAGE_ACCOUNT \
      -e AZURE_STORAGE_ACCESS_KEY devops:0.1 \
      /devops -pr 200 -wit -sem -v

```

Using the API
-------------
Call the API to get Pull request stats

```
abhinaba:~$ curl localhost:8080/pr?count=400

Reviewer Stats
              Trillian Astra 225 (56.2%) [############################################]
                Ford Prefect 140 (35.0%) [###############################-------------]
                 Arthur Dent 134 (33.5%) [#############################---------------]
              Slartibartfast 125 (31.2%) [###########################-----------------]
           Zaphod Beeblebrox 107 (26.8%) [####################------------------------]


Processed 400 pull-requests
```

Call the API to get workitem stats

```
abhinaba:~$ curl localhost:8080/wit
4884022: Deployments (Trillian Astra)
Done:2 InProgress:1 ToDo:3 Unknown:0
##=----

4884120: New SKU is onboarded (Ford Prefect)
Done:1 InProgress:0 ToDo:8 Unknown:0
#-----------

4669527: Reduce customer issues (Slartibartfast)
Done:3 InProgress:9 ToDo:14 Unknown:0
####============-------------------

3904063: Consumption tracking and Billing (Zaphod)
Done:5 InProgress:1 ToDo:16 Unknown:0
######=----------------------

3904108: Open source tools (Zaphod)
Done:5 InProgress:2 ToDo:14 Unknown:0
######==-------------------
```
