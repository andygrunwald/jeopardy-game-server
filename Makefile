.PHONY: install
install:
	go build

deploy-pi:
	GOARM=7 GOARCH=arm GOOS=linux go build && scp -r * pi@192.168.4.1:jeopardy-game-server/

deploy-pi-lan:
	GOARM=7 GOARCH=arm GOOS=linux go build && scp -r * pi@192.168.178.41:jeopardy-game-server/

.PHONY: run
run: install
	./jeopardy-game-server