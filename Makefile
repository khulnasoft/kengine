lint:
	golangci-lint run

fix-gci:
	gci -w -s standard -s default -s prefix(github.com/khulnasoft/kengine/v2/cmd) -s prefix(github.com/khulnasoft/kengine) --custom-order kengine.go admin.go

fix-gofumpt:
	gofumpt -w kengine.go admin.go modules/kenginetls/matchers.go

fix: fix-gci fix-gofumpt
