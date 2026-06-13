#!/bin/bash

run-notifier:
	cd ./notifier && go run main.go

run-listener:
	cd ./listener && go run main.go $(path)
