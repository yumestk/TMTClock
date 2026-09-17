NODE_MAJOR := 22

.PHONY: check-node dev-go dev-web dev test build clean

check-node:
	@node -v | grep -q v$(NODE_MAJOR) || (echo "wrong node: run 'nvm use' (.nvmrc)"; exit 1)

dev-go:
	go run ./cmd/tmtclock -no-browser

dev-web: check-node
	npm --prefix frontend run dev

dev: check-node
	go run ./cmd/tmtclock -no-browser & npm --prefix frontend run dev & wait

test:
	go test ./...

build: check-node
	rm -rf frontend/dist/assets frontend/dist/index.html
	npm --prefix frontend ci
	npm --prefix frontend run build
	CGO_ENABLED=0 go build -o tmtclock ./cmd/tmtclock

clean:
	rm -rf frontend/dist/assets frontend/dist/index.html tmtclock
