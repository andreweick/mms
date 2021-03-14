SAMPROFILE ?= edc-sam
REGION ?= us-east-1
PROJECT = $(shell basename $(CURDIR))
STACK_NAME ?= $(PROJECT)
STACK_NAME_PARAMS ?= $(STACK_NAME)-params
BUCKET ?= $(STACK_NAME)-$(SAMPROFILE)
METRIC_NAMESPACE ?= $(STACK_NAME)

BINARY ?= bin/$(NAME)
ZIP ?= bin/$(NAME).zip
MODULE ?= github.com/maeick/mms

export GOPRIVATE=git.missionfocus.com
export GOFLAGS=-mod=vendor

.PHONY: all build build-all update photograph audio video shuffle build-all clean test deploy destroy bucket validate

photograph: CMD = photograph;
photograph: NAME = MMSPhotograph
photograph:
	GOOS=linux CGO_ENABLED=0 go build -v -ldflags "-X main.version=$(CI_COMMIT_TAG)" -o $(BINARY) $(MODULE)/cmd/$(CMD)
	openssl sha1 $(BINARY) > $(BINARY).checksum
	zip -j $(ZIP) $(BINARY)

audio: CMD=audio
audio: NAME=MMSAudio
audio:
	GOOS=linux CGO_ENABLED=0 go build -v -ldflags "-X main.version=$(CI_COMMIT_TAG)" -o $(BINARY) $(MODULE)/cmd/$(CMD)
	openssl sha1 $(BINARY) > $(BINARY).checksum
	zip -j $(ZIP) $(BINARY)

shuffle: CMD=shuffle
shuffle: NAME=MMSShuffle
shuffle:
	GOOS=linux CGO_ENABLED=0 go build -v -ldflags "-X main.version=$(CI_COMMIT_TAG)" -o $(BINARY) $(MODULE)/cmd/$(CMD)
	openssl sha1 $(BINARY) > $(BINARY).checksum
	zip -j $(ZIP) $(BINARY)


build-all: mms audio shuffle

build:
	GOOS=linux CGO_ENABLED=0 go build -v -ldflags "-X main.version=$(CI_COMMIT_TAG)" -o $(BINARY) $(MODULE)/cmd/$(CMD)
	openssl sha1 $(BINARY) > $(BINARY).checksum
	zip -j $(ZIP) $(BINARY)

clean:
	@if [ -d ./bin/ ]; then rm -rf ./bin/; fi

test:
	cd shuffle/ ; go test -v ./...
	cd hello-world/ ; go test -v ./...

sam-build:
	@sam build

bucket:
	@aws s3 mb s3://$(BUCKET) \
		--profile $(SAMPROFILE) --region $(REGION)

validate:
	@cd shuffle ; go mod verify 
	@sam validate \
		--template .aws-sam/build/template.yaml 

package:
	# Package is no longer used, "deploy" does both packae and upload

deploy: validate bucket
	@sam deploy \
		--template .aws-sam/build/template.yaml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(BUCKET) \
		--s3-prefix $(STACK_NAME) \
		--no-fail-on-empty-changeset \
		--profile $(SAMPROFILE) --region $(REGION)

destroy:
	@aws cloudformation delete-stack \
		--stack-name $(STACK_NAME) \
		--profile $(SAMPROFILE) --region $(REGION)
	@aws s3 rb s3://$(BUCKET) --force \
		--profile $(SAMPROFILE) --region $(REGION)

