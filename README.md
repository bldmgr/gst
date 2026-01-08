# gst
Git Search Tools

A command-line tool written in Go for searching through git commit history and file contents.

## Features

- Search in commit messages
- Search in tracked file contents
- Interactive mode and single-query mode
- Displays detailed information about the last commit
- Works with any local git repository

## Prerequisites

- Go 1.16 or higher
- Git installed and accessible in your PATH

## Build and Run

### Build

To build the application, navigate to the project root directory and run:

```bash
go build -o gst main.go
```

This will create an executable named `gst` (or `gst.exe` on Windows).

### Run

You can run the built executable:

```bash
./gst
```

Or run directly using `go run`:

```bash
go run main.go
```

## Usage

### Command Line Arguments

- `-path`: Path to git repository (default: current directory)
- `-query`: Search query (if provided, runs a single search and exits)
- `-help`: Show help information

### Examples

Search for "bug fix" in the current repository:
```bash
./gst -query "bug fix"
```

Use a different repository:
```bash
./gst -path /path/to/repo
```

Start interactive mode in a specific repository:
```bash
./gst -path /path/to/repo
```

## How it works

The tool uses `git` command-line tools under the hood:
- `git log` for searching commit history
- `git grep` for searching file contents
- `git log -1` for retrieving the last commit details
