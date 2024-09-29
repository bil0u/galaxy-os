# Checking if .env file exists and if it does, include it
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

bot ?= default
version = dev
commit = $(shell git rev-parse --short HEAD)
source = cmd/*.go
target = /tmp/bin/${bot}
genScript = cmd/main.go
buildArgs = -ldflags "-X 'main.version=${version}' -X 'main.commit=${commit}'"
runArgs = --bot=${bot} --sync-commands --sync-roles --log-permissions

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
	go build ${buildArgs} -o=${target} ${source}
#   -> Include additional build steps, like TypeScript, SCSS or Tailwind compilation here...

run: build
	${target} ${runArgs}

## generate: generate go code
generate:
	WORKDIR=$(shell pwd) GENSCRIPT=${genScript} go generate ./...

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

deploy: target = /tmp/bin/linux_amd64/${bot}
deploy: version = $(shell git describe --tags --always --dirty)
deploy: buildArgs = -ldflags "-X 'main.commit=${commit}' -X 'main.version=${version}' -s"
deploy: runArgs = --bot=${bot} --sync-commands --sync-roles
deploy: confirm audit no-dirty
	GOOS=linux GOARCH=amd64 go build ${buildArgs} -o=${target} ${source}
	upx -5 ${target}
	# Include additional deployment steps here...

## deploy/hue: deploy the Hue bot to production
deploy/hue: bot = hue
deploy/hue: deploy

## deploy/kevin: deploy the Kevin bot to production
deploy/kevin: bot = kevin
deploy/kevin: deploy


