# deaddrop

A very simple tool to transfer files over HTTP using curl and wget. Intended to be used behind a reverse proxy.

## Usage
```bash
go build .
./deaddrop
```

## Transfer files
File with metadata: 
```bash
curl localhost:5050 -k -F@my_file.bin
curl localhost:5050 -k -F@my_file.bin --limit-rate 200K # 2 Gb/s = "2G", 3 Mb/s = "3M", 30 Kb/s = "30K"
```

Binary data: 
```bash
curl localhost:5050 -d@my_file.bin
```

Piped data: 

```bash
uname -a | curl localhost:5050 -d @-
```