# go-semantic-release

![go-semantic-release](https://github.com/Nightapes/go-semantic-release/workflows/Go/badge.svg)

## Release Types

| Type        |    Implemented     |      Git tag       |     Changelog      |      Release       |  Write access git  |     Api token      |
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
integrations:
  npm:
    enabled: true
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
  ## Optional, if you are not using github.com
  customUrl: <https://your.github>
  ## Optional, if you want to change the default tag prefix ("v")
  tagPrefix: ""
```

##### Gitlab 

You need to set the env `GITLAB_ACCESS_TOKEN` with an personal access token.

```yml
release: 'gitlab'
gitlab:
  repo: "<repositroyname>"  ## Example group/project
  ## Optional, if your not using gitlab.com
  customUrl: <https://your.gitlab>
  ## Optional, if you want to change the default tag prefix ("v")
  tagPrefix: ""
```

##### Git only 

Only via https at the moment. You need write access to your git repository


```yml
release: 'git'
git:
  email: "<email>" # Used for creating tag
  user: "<user>" : # Used for creating tag and pushing
  auth: "<token>" # Used for pushing, can be env "$GIT_TOKEN", will be replaced with env
  ## Optional, if you want to change the default tag prefix ("v")
  tagPrefix: ""
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
You can define hooks which run before or after the release. The shell commands will run in order, you can access the current release version via 
an environment variable `RELEASE_VERSION` 

```yml
hooks:
  preRelease:
  - name: echo $RELEASE_VERSION
  postRelease:
  - name: echo $RELEASE_VERSION
```

#### Integrations

Integrations are simple helpers to make integration with existing tools easier.
At the moment npm is supported, the integration will set the version before release to the `package.json`
```yml
integrations:
  npm:
    enabled: true
```

#### Changelog

Following variables and objects can be used for templates:

__Top level__

| Field                 | Type              | Description |
| --------              | ------            | -----       |
|   `Commits`           | string            | Fully rendered commit messages. This is left for backward compatibility. |
|   `CommitsContent`    | commitsContent    | Raw parsed commit data. Use this if you want to customize the output. |
|	`Version`           | string            | Next release version |
| 	`Now`               | time.Time         | Current time of generating changelog |
| 	`Backtick`          | string            | Backtick character |
| 	`HasDocker`         | bool              | If a docker repository is set in the config. |
| 	`HasDockerLatest`   | bool              | If `latest` image was uploaded |
| 	`DockerRepository`  | string            | Docker repository |

__commitsContent__

| Field                 | Type                           | Description |
| --------              | ------                         | -----       |
|   `Commits`           | map[string][]AnalyzedCommit    | Commits grouped by commit type |
|  	`BreakingChanges`   | []AnalyzedCommit               | Analyzed commit structure |
|  	`Order`             | []string                       | Ordered list of types |
|  	`HasURL`            | bool                           | If a URL is available for commits |
|  	`URL`               | string                         | URL for to the commit with {{hash}} suffix |

__AnalyzedCommit__

| Field                 | Type                  | Description |
| --------              | ------                | -----       |
|   `Commit`            | Commit                | Original GIT commit |
|  	`Tag`               | string                | Type of commit (e.g. feat, fix, ...) |
|  	`TagString`         | string                | Full name of the type |
|  	`Scope`             | bool                  | Scope value from the commit |
|  	`Subject`           | string                | URL for to the commit with {{hash}} suffix |
|   `MessageBlocks`     | map[string][]MessageBlock | Different sections of a message (e.g. body, footer etc.) |
|  `IsBreaking`         | bool                  | If this commit contains a breaking change |
|  `Print`              | bool                  | Should this commit be included in Changelog output |

__Commit__

| Field                 | Type                  | Description |
| --------              | ------                | -----       |
| `Message`             | string                | Original git commit message |
| `Author`              | string                | Name of the author |
| `Hash`                | string                | Commit hash value "|

__MessageBlock__

| Field                 | Type                  | Description |
| --------              | ------                | -----       |
| `Label`               | string                | Label for a block (optional). This will usually be a token used in a footer |
| `Content`             | string                | The parsed content of a block |

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

##### NPM

You can print a help text for a npm package

```yml
changelog:
  npm:
    name: ## Name of the npm package
    repository: ## Your docker repository, which is used for docker run
```


### Version

`go-semantic-release` has two modes for calculating the version: automatic or manual.

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

### Write changelog to file

This will write all changes beginning from the last git tag til HEAD to a changelog file. 
Default changelog file name if nothing is given via `--file`: `CHANGELOG.md`.
Note that per default the new changelog will be prepended to the existing file.
With `--max-file-size` a maximum sizes of the changelog file in megabytes can be specified.
If the size exceeds the limit, the current changelog file will be moved to a new file called `<filename>-<1-n>.<file extension>`. The new changelog will be written to the `<filename>`.
The default maximum file size limit is `10 megabytes`.

```bash
./go-semantic-release changelog --max-file-size 10
```

This will overwrite the given changelog file if its existing, if not it will be created.
```bash
./go-semantic-release changelog --overwrite
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

### Git Hooks

The `.githooks` folder contains a pre-commit script which has to be run before each commit.
To enable the hooks please add the `.githooks` folder to your core hooksPath via: `git config core.hooksPath .githooks`.
The following will be run in the pre-commit script:

- Formats all `*.go` files which are staged with `go fmt`
- Runs `go test` for the whole project
- Runs `golangci-lint` to check for linting errors