# View Annotated File

![Screenshot](/screenshot.png?raw=true "Screenshot")

## Install

```
go get github.com/loov/view-annotated-file
```

## Usage

```
go build -a -gcflags "-m -m -d=ssa/check_bce/debug" project 2> analysis.log
view-annotated-file analysis.log
```