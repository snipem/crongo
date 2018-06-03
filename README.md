# Crongo

A basic wrapper for cron

## Usage

    crongo run "ls -al | grep something"

Outcome of this call will be stored in the sqlite database under `$HOME/crongo.db`.

    crongo list all

Will display all runs like:

    ◉ 1 	2018-06-03T13:40:13Z	eintracht_transfers.py                            	Safely quitted webdriver                          	Traceback (most recent call last):                
                                                                                                                                            File "/home...                                  
    ◉ 0 	2018-06-03T13:45:02Z	/home/matze/bin/bikewatch.sh                      	                                                  	                                                  
    ◉ 0 	2018-06-03T13:45:04Z	temp_watch                                        	2018-06-03T13:45+0000,25.0                        	                                     

For only showing failed runs `crongo list failed` may be used. By default the number of displayed commands is 500. This can be overwritten by using `crongo list all 1000` or `crongo list failed 10` respectively.

## Cross compilation

    $ GOOS=linux GOARCH=amd64 go build
    $ file crongo
    crongo: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, with debug_info, not stripped
