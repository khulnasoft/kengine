lint:
	golangci-lint run

fix-gci:
	gci -w --all -s standard -s default -s prefix(github.com/khulnasoft/kengine/v2/cmd) -s prefix(github.com/khulnasoft/kengine) --custom-order

fix-gofumpt:
	gofumpt -w --recursive .

fix: fix-gci fix-gofumpt
