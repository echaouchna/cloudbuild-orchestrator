# cloudbuild-orchestrator (cork)

A simple orchestrator for Google Cloud Build triggers

## Usage

```sh
$ cork -h
Usage: cork [-exclude "<typeA,typeB,...>"] [-include "<type1,type2,...>"] [-no-fast-failing] [-reference <ref>] [-parallel <number>] <config_file>
  -exclude string
        Types to be excluded
  -include string
        Types to be included
  -no-fast-failing
        No fast failing
  -parallel int
        The number of parallel jobs (default 20)
  -reference string
        Reference to use for the build (default "develop")
  -version
        Version
```

or using docker

```sh
docker pull echaouchna/cork:latest
docker run --rm -it -v ~/.config/gcloud:/home/cork/.config/gcloud -v /path/to/config/file.yaml:/tmp/config.yaml echaouchna/cork:latest /tmp/config.yaml
```

## Example

Create config.yaml with the following content.

```yaml
name: demo application

steps:
  - name: terraform plan
    description: Run terraform plan for demo application
    trigger: demo-application-dev-tf-plan
    project-id: demo-app-6575
    tags: terraform

  - name: terraform apply
    description: Run terraform apply for demo application
    trigger: demo-application-dev-tf-apply
    project-id: demo-app-6575
    tags: terraform
    depends-on:
    - terraform plan
    manual: true

  - name: cicd trigger
    description: Run the cicd trigger for demo application
    trigger: cicd-develop-push-trigger
    project-id: demo-app-6575
    tags: cicd
    depends-on:
    - terraform plan

  - name: demo-application-deploy-dev
    description: Run deploy demo application
    trigger: demo-application-deploy-dev
    project-id: demo-app-6575
    tags: deploy
    depends-on:
    - cicd trigger
```

Run cork:

```sh
$ cork config.yaml
Using reference: develop
Fast failing: true
# demo application:
Nodes:
        terraform plan
        terraform apply
        cicd trigger
        demo-application-deploy-dev
Links:
        terraform plan -> <terraform apply>, <cicd trigger>
        cicd trigger -> <demo-application-deploy-dev>
Manual steps:
        terraform apply
[  RUNNING  ] [demo application/demo-application-dev-tf-plan] started
[  RUNNING  ] [demo application/demo-application-dev-tf-plan] triggered https://console.cloud.google.com/cloud-build/builds/buildid1?project=fakeproject
[  SUCCESS  ] [demo application/demo-application-dev-tf-plan] finished https://console.cloud.google.com/cloud-build/builds/buildid1?project=fakeproject
[  RUNNING  ] [demo application/cicd-develop-push-trigger] started
[  WAITING  ] [demo application/demo-application-dev-tf-apply] Please validate terraform plan to continue https://console.cloud.google.com/cloud-build/builds/buildid1?project=fakeproject (y/N):
[  RUNNING  ] [demo application/cicd-develop-push-trigger] triggered https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
[   SKIP    ] demo application cancelled by user
[  SUCCESS  ] [demo application/cicd-develop-push-trigger] finished https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
```

To include only some types (wildcards are accepted here):

```sh
$ cork -include "cicd,*deploy*" config.yaml
Using reference: develop
Fast failing: true
# demo application:
Nodes:
        terraform plan
        terraform apply
        cicd trigger
        demo-application-deploy-dev
Links:
        terraform plan -> <terraform apply>, <cicd trigger>
        cicd trigger -> <demo-application-deploy-dev>
Manual steps:
        terraform apply
[  RUNNING  ] [demo application/cicd-develop-push-trigger] started
[  RUNNING  ] [demo application/cicd-develop-push-trigger] triggered https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
[  SUCCESS  ] [demo application/cicd-develop-push-trigger] finished https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
[  RUNNING  ] [demo application/demo-application-deploy-dev] started
[  RUNNING  ] [demo application/demo-application-deploy-dev] triggered https://console.cloud.google.com/cloud-build/builds/buildid3?project=fakeproject
build failed
[   ERROR   ] [demo application/demo-application-deploy-dev] FAILURE https://console.cloud.google.com/cloud-build/builds/buildid3?project=fakeproject

```

To use a specific branch name (works also for hashcommits)

```sh
$ cork -reference feature/enhancements -include "cicd,deploy" config.yaml
Using reference: feature/enhancements
Fast failing: true
# demo application:
Nodes:
        terraform plan
        terraform apply
        cicd trigger
        demo-application-deploy-dev
Links:
        terraform plan -> <terraform apply>, <cicd trigger>
        cicd trigger -> <demo-application-deploy-dev>
Manual steps:
        terraform apply
[  RUNNING  ] [demo application/cicd-develop-push-trigger] started
[  RUNNING  ] [demo application/cicd-develop-push-trigger] triggered https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
[  SUCCESS  ] [demo application/cicd-develop-push-trigger] finished https://console.cloud.google.com/cloud-build/builds/buildid2?project=fakeproject
...
```
