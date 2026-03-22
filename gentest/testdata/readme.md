## Readme

blah blah

### Key Capabilities

blah blah

### Requirements

requirements

```sh
# macOS
brew install pkl
```

### Usage

#### As a standalone generator 

First let's set up a golang module

```go.mod
module nhatp.com/go/example
go 1.24
```

using the minimum configuration

```pkl
// your pkl config here
```

source code

```go
// file: main.go
package main
```

expected generated code

```go
// golden-file: gen.go
package main
```
