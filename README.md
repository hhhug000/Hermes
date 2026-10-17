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

## Want to try?

There is no public test server, but you can host one youself!
Don't worry this is super simple.
Just clone the repo, and run the code!

This example requires that you have golang installed, and you are on linux, most commands will work on windows, but CP might not.
To do that on windows just copy the example config file to the same folder and rename it to hermes.conf

```
git clone https://github.com/hhhug000/Hermes.git
cd Hermes
cp hermes.conf.example hermes.conf
go run .
```

Then in a new terminal run the following command to connect

```
ssh localhost -p 2222
```

## Not working?

If Hermes is not working, try making sure you have at least 1 default SSH key installed, as Hermes uses this for user identification.
