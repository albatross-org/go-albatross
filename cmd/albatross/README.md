# `albatross`
`albatross` is a command line tool for working with Albatross stores. This makes it a powerful tool for managing networked thoughts, ideas and information.

- [`albatross`](#albatross)
  - [Setup](#setup)
    - [Global Configuration](#global-configuration)
    - [Store-Level Configuration](#store-level-configuration)
    - [Example](#example)
    - [Using Git](#using-git)
  - [Usage](#usage)
  - [Implementation](#implementation)

## Setup
`albatross` requires two configurations. At the moment, this has to be set up manually but in future it would be nice to have this done automatically by a command such as `albatross initialise`.

### Global Configuration
This file should be placed into the `~/.config/albatross` directory, named `config.yaml`:

```
default:
    path: "/path/to/albatross/store"
```

This gives names to Albatross stores. It would be inconvient to write out the full path everytime when using the command, so this specifies a shorthand:

```sh
$ albatross ...
# Uses the "default" Albatross store

$ albatross --store "phd"
# Uses the "phd" Albatross store 
```

If no store is explicitely specified, `default` is used.

### Store-Level Configuration
The directory the global config points to should be formatted like so:

```
config.yaml - Config file
entries/ - Where the entries live
templates/ - Templates, see albatross create --help for more info
```

`config.yaml` can contain the following:

```yaml
dates:
  format: "2006-01-02 15:04PM" # Go date format used here.

tags:
  prefix-builtin: "@!"
  prefix-custom: "@?"

encryption:
  public-key: "/path/to/public/pgp/key"
  private-key: "/path/to/private/pgp/key"
```

Though they are all optional.

### Example
Here's an example of a configuration.

> `~/.config/albatross/config.yaml`
```
default:
    path: "/home/olly/.local/share/albatross/default"

testing:
    path: "/home/olly/.local/share/albatross/testing/testing"
```

> `~/home/olly/.local/share/albatross/default/config.yaml`
```
dates:
  format: "2006-01-02 15:04" # Go date format used here.

tags:
  prefix-builtin: "@!"
  prefix-custom: "@?"

encryption:
  public-key: "/home/olly/.config/albatross/keys/public.key"
  private-key: "/home/olly/.config/albatross/keys/private.key"
```

### Using Git
If you initialise the `entries/` folder as a Git repository, you can access version control using `albatross git`. It will also automatically track changes.

See `albatross git --help` for more information.

## Usage
See

```
$ albatross --help
```

## Implementation
`albatross` uses [`go-albatross`](https://github.com/albatross-org/go-albatross) in order to work with Albatross stores.