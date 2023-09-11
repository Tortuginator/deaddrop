# deaddrop

A very simple tool to transfer files over HTTP using curl and wget. Intended to be used behind a reverse proxy.

## Usage
```bash
go build .
./deaddrop
```

## Transfer files
File with metadata: `curl localhost:5050 -F@my_file.bin`
Binary data: `curl localhost:5050 -d@my_file.bin`
Piped data: `curl localhost:5050 -d @- | uname -a`