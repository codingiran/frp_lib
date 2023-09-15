export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on
LDFLAGS := -s -w
GCFLAGS := -l -B

all: fmt build

#build: frpc-lib-dll

# compile assets into binary file
file:
	rm -rf ./assets/frps/static/*
	rm -rf ./assets/frpc/static/*
	cp -rf ./web/frps/dist/* ./assets/frps/static
	cp -rf ./web/frpc/dist/* ./assets/frpc/static

fmt:
	go fmt ./...

fmt-more:
	gofumpt -l -w .

gci:
	gci write -s standard -s default -s "prefix(github.com/Bellegar/frp_lib/)" ./

vet:
	go vet ./...

frps:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/frps ./cmd/frps

frpc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/frpc ./cmd/frpc

frpc-lib-static:
	env CGO_ENABLED=1 go build -trimpath -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-archive -o bin/static/ ./cmd/libfrpc

frpc-lib-static-macos:
	env GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-archive -o bin/x86_64/libfrpc.a ./cmd/libfrpc
	env GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-archive -o bin/arm64/libfrpc.a ./cmd/libfrpc
	mkdir -p bin/arm64_fat
	lipo -create bin/x86_64/libfrpc.a bin/arm64/libfrpc.a -output bin/arm64_fat/libfrpc.a
	cp bin/x86_64/libfrpc.h bin/arm64_fat/libfrpc.h

frpc-lib-unix-x64:
	env GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-shared -o bin/x86_64/ ./cmd/libfrpc

frpc-lib-unix-arm64:
	env GOARCH=arm64 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-shared -o bin/arm64/ ./cmd/libfrpc

frpc-lib-windows64:
	env GOARCH=amd64 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-shared -o bin/x86_64/frpc.dll ./cmd/libfrpc
	cmd /C dll2lib.bat 64 bin/x86_64/frpc.dll
	mv frpc.lib bin/x86_64/frpc.lib

frpc-lib-windows32:
	env GOARCH=386 CGO_ENABLED=1 go build -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS) -extldflags=-static" -buildmode=c-shared -o bin/x86/frpc.dll ./cmd/libfrpc
	cmd /C dll2lib.bat 32 bin/x86/frpc.dll
	mv frpc.lib bin/x86/frpc.lib

frpc-lib-android:
	go get golang.org/x/mobile/cmd/gobind
	go get golang.org/x/mobile/cmd/gomobile
	mkdir -p bin/android
	gomobile init
	gomobile bind -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS)" -v -o bin/android/libfrp.aar -target=android ./cmd/libfrpc_mobile

frpc-lib-ios:
	go get golang.org/x/mobile/cmd/gobind
	go get golang.org/x/mobile/cmd/gomobile
	mkdir -p bin/ios
	gomobile init
	gomobile bind -trimpath -gcflags=all="$(GCFLAGS)" -ldflags "$(LDFLAGS)" -v -o bin/ios/libfrpc.xcframework -target=ios ./cmd/libfrpc_mobile

test: gotest

gotest:
	go test -v --cover ./assets/...
	go test -v --cover ./cmd/...
	go test -v --cover ./client/...
	go test -v --cover ./server/...
	go test -v --cover ./pkg/...

e2e:
	./hack/run-e2e.sh

e2e-trace:
	DEBUG=true LOG_LEVEL=trace ./hack/run-e2e.sh

e2e-compatibility-last-frpc:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPC_PATH="`pwd`/lastversion/frpc" ./hack/run-e2e.sh
	rm -r ./lastversion

e2e-compatibility-last-frps:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPS_PATH="`pwd`/lastversion/frps" ./hack/run-e2e.sh
	rm -r ./lastversion

alltest: vet gotest e2e

clean:
	rm -f ./bin/frpc
	rm -f ./bin/frps
	rm -rf ./lastversion
