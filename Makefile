watchman.test:
	@cat testwatchconfig.json | jq -c . | watchman -j -p --no-pretty > output.json

mstack.watch:
	@go run src/go/tools/mstack/main.go watch


watchman.query:
	@cat watchmanquery.json | jq -c . | watchman -j -p > output.json

vendor:
	go mod vendor

ark.build:
	go build -o ./src/go/bin/arkcliv2 ./src/go/tools/arkcliv2/*.go

# execute with no debugger
ark.debug.build:
	# build without compiler optimization
	go build -gcflags="all=-N -l" -o ./src/go/bin/ark.debug ./src/go/tools/arkcliv2/*.go

ark.debug.install:
	install ./src/go/bin/ark.debug /usr/local/bin/ark.debug

ark.debug.server-run: 
	ark run server

#1: delve locally
ark.debug.local:
	dlv debug --help | head -13
	dvl debug ./src/go/tools/arkcliv2/main.go

#2: delve debug test locally
ark.debug.test:
	dlv test --help | head -12
	dlv test ./src/go/tools/arkcliv2/main.go

#3: locally running process
ark.debug.attach:
	dlv attach --help | head -12
	pgrep ark.debug
	dlv attach $(shell pgrep ark.debug)

#4: delve exec locally
ark.debug.exec:
	dlv exec --help | head -13
	dlv exec ark.debug -- something

#5: delve debug trace
ark.debug.trace:
	dvl trace ./src/go/tools/arkcliv2/main.go -- server run

#6: connecting to a remote server
# explanation of flags:
# --accept-multiclient               Allows a headless server to accept multiple client connections.
# --headless                         Run debug server only, in headless mode.

ark.debug.server:
	dlv debug ./src/go/tools/arkcliv2/main.go --accept-multiclient --headless localhost:4000 --api-version 2 -- server run

ark.debug.kill:
	# "-f" searches for matching expression, "-l" list PID
	pkill -f -l ark.debug

ark.clean:
	rm -rf ./src/go/bin
	install ./src/go/bin/arkcliv2 $$HOME/.local/bin/ark

.PHONY: ark.build ark.clean vendor watchman.test watchman.query watchman.watch
