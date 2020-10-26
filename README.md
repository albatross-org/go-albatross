# `go-albatross`
`go-albatross` is a Go implementation of the Albatross Core API.

> To use `go-albatross` from the command line, you're probably looking for the [`albatross`](https://github.com/albatross-org/go-albatross/cmd/albatross) command line tool.

### Contributing
This code, although functional, is still in it's early stages.

### Project Structure

```
go-albatross/
├── cmd/
│   └── albatross/ # The albatross command line tool.
│
├── pkg/
│   └── core/ # An implementation of the core API. Although it's called core, typically it would be imported as "albatross".
│
├── entries/ # An entries parser. This package deals with processing an entries folder containing the extended markdown syntax
│            # and turning it into an internal graph.
│            # It defines an Entry struct and an Entries data structure which allowed for quick searching and processing.
│
├── encryption/ # This package deals with providing encryption functionality. This will use OpenPGP and provide an simple API
│               # for encrypting and decrpyting folders with public and private keys.
│
├── version.go # Holds version information.
├── doc.go # Go doc file.
│
├── README.md # Readme.
│
├── go.mod # Go modules.
└── go.sum # Go modules.
```