# Crongo

A basic wrapper for cron

## Usage

    crongo run "ls -al | grep something"

Outcome of this call will be stored in the sqlite database under `$HOME/crongo.db`.

## Cross compilation

    $ GOOS=linux GOARCH=amd64 go build
    $ file crongo
    crongo: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, with debug_info, not stripped
