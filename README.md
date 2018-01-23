# View Annotated File

![Screenshot](/screenshot.png?raw=true "Screenshot")

## Install

```
go get github.com/loov/view-annotated-file
```

## Usage go1.9

```
go build -a -gcflags "-m -m -d=ssa/check_bce/debug" project 2> analysis.log
view-annotated-file analysis.log
```

## Usage go1.10

```
go build -a -gcflags "all=-m -m -d=ssa/check_bce/debug" project 2> analysis.log
view-annotated-file analysis.log
```
