default: binary

binary:
	GOOS=linux GOARCH=amd64 go build -o proxy

container: container-clean binary
	docker build -t proxy -f Dockerfile .
	docker run -tdi -p 1080:1080 --name proxy proxy

container-clean:
	docker rm -f proxy || echo
	docker rmi proxy || echo

clean:
	rm -f proxy
