# View Annotated File

![Screenshot](/screenshot.png?raw=true "Screenshot")

## Install

```
go get github.com/loov/view-annotated-file
```

## Usage

```
go build -a -gcflags "-m -m" project 2> escape.analysis
view-annotated-file escape.analysis
```