# Crongo

[![Build Status](https://travis-ci.org/snipem/web-to-feed.svg?branch=master)](https://travis-ci.org/snipem/web-to-feed)
[![Coverage Status](https://coveralls.io/repos/github/snipem/crongo/badge.svg?branch=master)](https://coveralls.io/github/snipem/crongo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/snipem/crongo)](https://goreportcard.com/report/github.com/snipem/crongo)

A basic wrapper for cron

## Usage

### Run

    crongo run "ls -al | grep something"

Outcome of this call will be stored in the sqlite database under `$HOME/crongo.db`.

### List

    crongo list all

Will display all runs like:

    CODE 	ID	DATE                          	CMD                    	STDOUT                                            	STDERR
    ◉ 0  	1 	2018-05-30 21:29:45 +0200 CEST	uptime                 	21:29  up 5 days,  3:10, 8 users, load averages...
    ◉ 0  	2 	2018-05-30 22:02:16 +0200 CEST	uptime                 	22:01  up 5 days,  3:42, 8 users, load averages...
    ◉ 0  	3 	2018-05-31 10:07:26 +0200 CEST	uptime                 	10:07  up 5 days, 15:48, 7 users, load averages...
    ◉ 0  	4 	2018-05-31 11:14:12 +0200 CEST	uptime                 	11:13  up 5 days, 16:54, 7 users, load averages...
    ◉ 0  	5 	2018-05-31 11:14:59 +0200 CEST	ls                     	README.md
                                                                        beatle
                                                                        commute-tube.log
                                                                        crongo.db
                                                                        crongo....

For only showing failed runs `crongo list failed` may be used. By default the number of displayed commands is 500. This can be overwritten by using `crongo list all 1000` or `crongo list failed 10` respectively.

### Details

For getting the whole stdout and stderr you can run

    crongo id 5

## Cross compilation

    $ GOOS=linux GOARCH=amd64 go build
    $ file crongo
    crongo: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, with debug_info, not stripped
