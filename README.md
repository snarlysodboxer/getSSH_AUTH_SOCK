# getSSH_AUTH_SOCK

## Find a valid ssh-agent socket file and return the path to it, remove stale socket files.

## Explain
You use SSH Agent Forwarding with Tmux on a remote server. You loose your SSH tunnel, when you reconnect, you want the SSH_AUTH_SOCK environment variable to be automatically reset for you.
getSSH_AUTH_SOCK will look for a valid socket file and output the path to it on `stdout`. If it finds more than one socket file, it'll delete any that don't respond to `ssh-add -l` within 2 seconds. If it encounters any errors, it prints them to `stderr` and exit's 1.

### Installation
Build the app and put it in somewhere like `/usr/local/bin`.
```bash
go build && sudo cp getSSH_AUTH_SOCK /usr/local/bin
```

In your `.bashrc` or similar:

```bash
reset_ssh () {
    if SOCKET=$(getSSH_AUTH_SOCK); then
        export SSH_AUTH_SOCK="$SOCKET"
    fi
}
export PROMPT_COMMAND='reset_ssh'
```

### Usage
After re-attaching to your Tmux session, simply hit enter to get a new Bash prompt, and Agent Forwarding will work again.

### Advanced usage
```bash
$ resetSSH_AUTH_SOCK --help
Usage of resetSSH_AUTH_SOCK:
  -regex string
        Regular expression to match subdirectory names against (default "ssh-.*")
  -rootDir string
        Directory to scan for SSH socket files (default "/tmp")
$
```

