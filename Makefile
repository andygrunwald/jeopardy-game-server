.PHONY: install
install:
	go build

.PHONY: run
run: install
	./jeopardy-game-server
