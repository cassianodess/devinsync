ifneq (,$(wildcard ./listener/.env))
include ./listener/.env
export
endif

run-notifier:
	cd ./notifier && go run main.go

run-host:
	cd ./listener && go run main.go -target=host -path=$(path)

run-guest:
	cd ./listener && go run main.go -target=guest -room=$(room)
