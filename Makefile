.PHONY: build

build:
	@go build -o wcg github.com/joelspeed/webhook-certificate-generator/cmd/webhook-certificate-generator
