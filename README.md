# Hermes
An SSH powered file transfer system, named after the messenger of the greek gods

## Features

- Runs over ssh, nothing needs to be installed on any device
- Authentication based on SSH key fingerprints
- Access control, only share files with people you want to see
- Pipe in uploads with `cat file | ssh SERVER -p PORT send file`
- Streaming downloads with `ssh SERVER -p PORT download fileid > file`

## Architecture

Hermes is built in Go.
It uses the wish framework to work over SSH,
Bubbletea for the TUI
SQLite for the database
Normal files for storage

## Config

Just make a config file like this in the running directory and fill in your ip and port:

```
host = localhost
port = 2222
download_dir = .
```

Make sure to name it `hermes.conf`