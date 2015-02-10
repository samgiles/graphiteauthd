default: bin/graphiteauthd

bin/:
	mkdir -p bin/

bin/graphiteauthd: bin/ main.go
	go build main.go && mv main bin/graphiteauthd

clean:
	rm -rf bin/

test:
	go test

bench:
	go test -bench . -benchmem
