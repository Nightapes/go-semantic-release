# go-semantic-release

![go-semantic-release](https://github.com/Nightapes/go-semantic-release/workflows/Go/badge.svg)

## Release Types

| Type        |    Implemendet     |      Git tag       |     Changelog      |      Release       |  Write access git  |     Api token      |
| ----------- | :----------------: | :----------------: | :----------------: | :----------------: | :----------------: | :----------------: |
| `github`    | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |                    | :white_check_mark: |
| `gitlab`    | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |                    | :white_check_mark: |
| `git`       | :white_check_mark: | :white_check_mark: |                    |                    | :white_check_mark: |                    |
| `bitbucket` |    Comming soon    | :white_check_mark: |                    |                    | :white_check_mark: |                    |


## Supported CI Pipelines

* Github Actions
* Gitlab CI
* Travis CI
* Custom CI, set enviroment `CI=true`

## Download

You can download the newest version under [releases](https://github.com/Nightapes/go-semantic-release/releases)

or

you can use a Docker image

`docker pull nightapes/go-semantic-release:<VERSION>` or `docker pull docker.pkg.github.com/nightapes/go-semantic-release/go-semantic-release:<VERSION>`



## How to use

`go-semantic-release` config file 
Create a file with the name `.release.yml` or anything else, but you need to set to every command `-c <your config file>`

### Example config

```yml
commitFormat: angular
branch:
  master: release
release: 'github'
github:
  repo: "go-semantic-release"
  user: "nightapes"
assets:
  - name: ./build/go-semantic-release
    compress: false
  - name: ./build/go-semantic-release.exe
    compress: false
hooks:
  preRelease:
    - name: echo $RELEASE_VERSION
  postRelease:
    - name: echo $RELEASE_VERSION
```

#### CommitFormat

Supported formats:

* [angular](https://github.com/angular/angular/blob/master/CONTRIBUTING.md#commit-message-format)

    ```yml
    commitFormat: angular
    ```

* [conventional](https://www.conventionalcommits.org/en/v1.0.0/#summaryhttps://www.conventionalcommits.org/en/v1.0.0/#summary)

    ```yml
    commitFormat: conventional
    ```


#### Branch

You can define which kind of release should be created for different branches. 

Supported release kinds: 

* `release` -> `v1.0.0`
* `rc` -> `v1.0.0-rc.0`
* `beta` -> `v1.0.0-beta.0`
* `alpha` -> `v1.0.0-alpha.0`

Add a branch config to your config

```yml
branch:
  <branch-name>: <kind>
```

#### Release

At the moment we support releases to gitlab and github.

##### Github 

You need to set the env `GITHUB_TOKEN` with an access token.

```yml
release: 'github'
github:
  user: "<user/group"
  repo: "<repositroyname>"
  ## Optional, if your not using github.com
  customUrl: <https://your.github>
```

##### Gitlab 

You need to set the env `GITLAB_ACCESS_TOKEN` with an personal access token.


```yml
release: 'gitlab'
gitlab:
  repo: "<repositroyname>"  ## Example group/project
  ## Optional, if your not using gitlab.com
  customUrl: <https://your.gitlab>
```

##### Git only 

Only via https at the moment. You need write access to your git repository


```yml
release: 'git'
git:
  email: "<email>" # Used for creating tag
  user: "<user>" : # Used for creating tag and pushing
  auth: "<token>" # Used for pushing, can be env "$GIT_TOKEN", will be replaced with env
```


#### Assets

You can upload assets to a release

Support for gitlab and github.
If you want, you can let the file be compressed before uploading 

```yml
assets:
  - name: ./build/go-semantic-release
    compress: false
```

#### Hooks

Hooks will run when calling `release`. Hooks run only if a release will be triggered.  

#### Changelog

Following variables can be used for templates:
*   `Commits` string
*	`Version` string
* 	`Now` time.Time
* 	`Backtick` string
* 	`HasDocker` bool
* 	`HasDockerLatest` bool
* 	`DockerRepository` string

```yml
changelog:
  printAll: false ## Print all valid commits to changelog
  title: "v{{.Version}} ({{.Now.Format "2006-01-02"}})" ## Used for releases (go template)
  templatePath: "./examples/changelog.tmpl"    ## Path to a template file (go template)
```

##### Docker 

You can print a help text for a docker image

```yml
changelog:
  docker: 
    latest: false ## If you uploaded a latest image
    repository: ## Your docker repository, which is used for docker run
```

### Version

`go-semantic-release` has two modes for calcualting the version: automatic or manual.

#### Automatic

Version will be calculated on the `next` or `release` command

#### Manual

If you don't want that `go-semantic-release` is calculating the version from the commits, you can set the version by hand with
following command:

```bash
./go-semantic-release set 1.1.1
```

### Print version

Print the next version, can be used to add version to your program

```bash
./go-semantic-release next // show next version (calculated by new commits since last version)
./go-semantic-release last // show last released version 
```
Example with go-lang

```bash
go build -ldflags "--X main.version=`./go-semantic-release next`"
```

### Create release 

```bash
./go-semantic-release release 
```



## Build from source

```bash
go build ./cmd/go-semantic-release/
```

### Testing

```bash
go test ./... 
```

### Linting

```
curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.16.0
golangci-lint run ./...
```
