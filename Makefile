

# change to your device to test with
dev/air:
	air -build.args_bin "--serial ${SERIAL}"

dev:
	templ generate && tailwindcss -i ./internal/web/static/css/input.css -o ./internal/web/static/css/dist/style.css 2>/dev/null && go run ./cmd/noreza/main.go --serial "${SERIAL}"

profile:
	templ generate && \
		tailwindcss -i ./internal/web/static/css/input.css -o ./internal/web/static/css/dist/style.css 2>/dev/null && \
		go run ./cmd/noreza/main.go --serial "${SERIAL}" --cpuprofile cpu.prof --memprofile mem.prof

.PHONY: dev dev/air profile