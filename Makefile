default: bin/graphiteauthd

get-deps:
	go get github.com/mgutz/ansi

bin/:
	mkdir -p bin/

bin/graphiteauthd: bin/ main.go get-deps
	go build main.go && mv main bin/graphiteauthd

clean:
	rm -rf bin/

test:
	go test

bench:
	go test -bench . -benchmem
