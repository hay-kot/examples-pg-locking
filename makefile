.PHONE: run
run:
	docker stop example.pg.locking || true
	docker run --name example.pg.locking \
						--rm -p 5432:5432 \
						-e POSTGRES_USER=postgres \
						-e POSTGRES_PASSWORD=postgres \
						-e POSTGRES_DB=example_db \
						-d postgres:13-alpine
	go run main.go


