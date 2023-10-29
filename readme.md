# deaddrop

A very simple tool to transfer files over HTTP using LotL techniques. Intended to be used behind a reverse proxy providing TLS encrypytion.

## Usage
```bash
go build .
./deaddrop
```

## (Curl) Transfer files 
### Small files < 100Mb

Using the following commands with large files may cause a Out-of-memory exception in curl on your machine.

File upload with metadata using curl: 
```bash
curl localhost:5050 -k -F file=@my_file.bin 
```
Binary data upload using curl: 
```bash
curl localhost:5050 -d@my_file.bin
curl localhost:5050 -T my_file.bin --progress-bar 
```

Piped data upload using curl: 

```bash
uname -a | curl localhost:5050 -d @-
```

You may use `--limit-rate 200K` to limit the upload speed with curl. [Futher information](https://everything.curl.dev/usingcurl/transfers/rate-limiting).

### Large files >= 100Mb

The following command leverages request streaming and does not cause Out-of-memory exceptions:
```bash
curl localhost:5050 -T my_file.bin --progress-bar 
```

## (wget) Transfer files
```bash
wget --post-file=".\my_file.bin"  http://localhost:5050
```
## (CertReq) Transfer files

The windows binary CertReq can be used to upload files using the following command:
```bash
CertReq -Post -config  "http://localhost:5050/" my_file.bin
```
More information: [LOLBAS](https://lolbas-project.github.io/lolbas/Binaries/Certreq/)

## (Powershell) Transfer files

The following command uploads files using powershell:
```powershell
Invoke-RestMethod -Uri "http://localhost:5050" -Method Post -InFile ".\my_file.bin"
```