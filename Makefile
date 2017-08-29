.PHONEY: test
test:
	go test -cover ./...

.PHONEY: bench
bench:
	go test -cover -bench . -benchmem -run 'Benchmark.*' ./...

.PHONEY: vet
vet:
	go vet ./...

# Requires graphviz
deps.svg: *.go
	(   echo "digraph G {"; \
	go list -f '{{range .Imports}}{{printf "\t%q -> %q;\n" $$.ImportPath .}}{{end}}' \
		$$(go list -f '{{join .Deps " "}}' .) .; \
	echo "}"; \
	) | dot -Tsvg -o $@
