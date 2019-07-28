# go-semantic-release

## Release Types

| Type      | Git tag               | Changelog             | Release           |  Write access git | Api token           |
|---        |:---:                  |:---:                  |:---:              |:---:              |:---:                |
| `git`     | :white_check_mark:    |                       |                   | :white_check_mark:|                     |
| `github`  | :white_check_mark:	| :white_check_mark:    | :white_check_mark:|                   | :white_check_mark:  |
| `gitlab`  | :white_check_mark:	| :white_check_mark:    | :white_check_mark:|                   | :white_check_mark:  |



## Build

`go build ./cmd/go-semantic-release/`

## Run

Print the next version

`./go-semantic-release version next`

Set a version

`./go-semantic-release version set v1.1.1`
