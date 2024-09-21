# Checking if .env file exists and if it does, include it
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

bot ?= none
target = ./bots/${bot}

# =======
# HELPERS
# =======

.PHONY: help confirm no-dirty

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
	
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

no-dirty:
	@test -z "$(shell git status --porcelain)"


# ===============
# QUALITY CONTROL
# ===============

.PHONY: audit test test/cover

## audit: run quality control checks
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)" 
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out


# ===========
# DEVELOPMENT
# ===========

.PHONY: tidy push build run build/hue build/kevin run/hue run/kevin

## tidy: tidy modfiles and format .go files
tidy:
	go mod tidy -v
	go fmt ./...

## push: push changes to the remote Git repository
push: confirm audit no-dirty
	git push

build: tidy
	go build -o=/tmp/bin/${bot} ${target}
#   -> Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...

run: build
	/tmp/bin/${bot}

## build/hue: build the Hue bot
build/hue: bot = hue
build/hue: build

## build/kevin: build the Kevin bot
build/kevin: bot = kevin
build/kevin: build

## run/hue: run the Hue bot
run/hue: bot = hue
run/hue: run

## run/kevin: run the Kevin bot
run/kevin: bot = kevin
run/kevin: run

# ==========
# PRODUCTION
# ==========

.PHONY: deploy deploy/hue deploy/kevin

deploy: confirm audit no-dirty
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/${bot} ${target}
	upx -5 /tmp/bin/linux_amd64/${bot}
	# Include additional deployment steps here...

## deploy/hue: deploy the Hue bot to production
deploy/hue: bot = hue
deploy/hue: deploy

## deploy/kevin: deploy the Kevin bot to production
deploy/kevin: bot = kevin
deploy/kevin: deploy


