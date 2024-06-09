# Params to test
SERVER_PORT=8080
TEMP_FILE=/tmp/metrics.json
DATABASE_DSN=postgres://postgres:metrics@localhost:5432/metrics?sslmode=disable


all: build vet iter1 iter2 iter3 iter4 iter5 iter6 iter7 iter8 iter9 iter10 iter11 iter12 iter13 iter14
test: build vet iter11 iter12
build:
	go build -C cmd/agent -o agent
	go build -C cmd/server -o server
	go build -C cmd/staticlint -o staticlint
vet:
	goimports -local "github.com/arxon31/metrics-collector" -w /Users/arxon31/go/src/github.com/arxon31/metrics-collector/
	go vet -vettool=statictest-darwin-arm64 ./...

iter1:
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$$ \
                -binary-path=cmd/server/server
iter2:
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2[AB]*$ \
                -source-path=. \
                -agent-binary-path=cmd/agent/agent
iter3:
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration3[AB]*$ \
            	-source-path=. \
            	-agent-binary-path=cmd/agent/agent \
            	-binary-path=cmd/server/server
iter4:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration4$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter5:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration5$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter6:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration6$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter7:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration7$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter8:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration8$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter9:
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration9$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -file-storage-path=$(TEMP_FILE) \
            -server-port=$(SERVER_PORT) \
            -source-path=.
iter10:
	docker run --rm -d --name=metricsDB -e=POSTGRES_PASSWORD=metrics -e=POSTGRES_DB=metrics -p 5432:5432 postgres
	sleep 2
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration10[AB]$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn=$(DATABASE_DSN) \
            -server-port=$(SERVER_PORT) \
            -source-path=.
	docker stop metricsDB
iter11:
	docker run --rm -d --name=metricsDB -e=POSTGRES_PASSWORD=metrics -e=POSTGRES_DB=metrics -p 5432:5432 postgres
	sleep 2
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration11$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn=$(DATABASE_DSN) \
		-server-port=$(SERVER_PORT) \
		-source-path=.
	docker stop metricsDB
iter12:
	docker run --rm -d --name=metricsDB -e=POSTGRES_PASSWORD=metrics -e=POSTGRES_DB=metrics -p 5432:5432 postgres
	sleep 2
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration12$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn=$(DATABASE_DSN) \
		-server-port=$(SERVER_PORT) \
		-source-path=.
	docker stop metricsDB
iter13:
	docker run --rm -d --name=metricsDB -e=POSTGRES_PASSWORD=metrics -e=POSTGRES_DB=metrics -p 5432:5432 postgres
	sleep 2
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration13$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn=$(DATABASE_DSN) \
		-server-port=$(SERVER_PORT) \
		-source-path=.
	docker stop metricsDB
iter14:
	docker run --rm -d --name=metricsDB -e=POSTGRES_PASSWORD=metrics -e=POSTGRES_DB=metrics -p 5432:5432 postgres
	sleep 2
	SERVER_PORT=$(SERVER_PORT)
	ADDRESS="localhost:$(SERVER_PORT)"
	TEMP_FILE=$(TEMP_FILE)
	./metricstest-darwin-arm64 -test.v -test.run=^TestIteration14$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn=$(DATABASE_DSN) \
		-key="$(TEMP_FILE)" \
		-server-port=$(SERVER_PORT) \
		-source-path=.
		docker stop metricsDB
