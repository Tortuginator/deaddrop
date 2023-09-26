# deaddrop

A very simple tool to transfer files over HTTP using curl. Intended to be used behind a reverse proxy providing TLS encrypytion.

## Usage
```bash
go build .
./deaddrop
```

## Transfer files
File upload with metadata using curl: 
```bash
curl localhost:5050 -k -F file=@my_file.bin 
```
You may use `--limit-rate 200K` to limit the upload speed with curl. [Futher information](https://everything.curl.dev/usingcurl/transfers/rate-limiting).

Binary data upload using curl: 
```bash
curl localhost:5050 -d@my_file.bin
curl localhost:5050 -T my_file.bim --progress-bar 

```

Piped data upload using curl: 

```bash
uname -a | curl localhost:5050 -d @-
```