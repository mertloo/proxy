default: binary

binary:
	GOOS=linux GOARCH=amd64 CGO_ENABLE=0 go build -o proxy ./cmd

clean:
	rm -f proxy
