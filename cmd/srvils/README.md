# srvils

## Build information

The srvils executable includes a web user interface. This is embedded into the executable using
https://github.com/mjibson/esc.

The webui must be built using `yarn build`. See the webui/README.mb for instructions.

When the web ui has been built, run `go generate ./cmd/srvils/httpapi.go` to generate the webui.go
containing the web assets. webui.go is not under source control.

### Generated code

Install stringer and esc (run outside the project folder):
`go get -u golang.org/x/tools/cmd/stringer github.com/mjibson/esc`

Generate with: `go generate ./cmd/srvils/httpapi.go`
