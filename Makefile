

run-server: 
	@go run ./cmd/

run-web: 
	@go run ./cmd/web/

run-frontend: 
	@npm run --prefix frontend dev