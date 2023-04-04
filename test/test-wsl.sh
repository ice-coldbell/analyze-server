# /bin/bash
go clean -testcache
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o cover.html
rm coverage.out
cmd.exe /C $(wslpath -w $(realpath cover.html))
sleep 1
rm cover.html