# wg-ddns
WireGuard DDNS Endpoint

## Usage

Install the binary using the following command:

### A/AAAA Record

```shell
go install github.com/koho/wg-ddns/cmd/wg-ip@latest
```

### SVCB Record

```shell
go install github.com/koho/wg-ddns/cmd/wg-svcb@latest
```

### Windows

Install the background service:

```shell
sc create wg-ddns binPath= "%GOPATH%\bin\wg-ip.exe -i YOUR_INTERFACE" DisplayName= "WireGuard DDNS" start= auto
```

Start the service:

```shell
net start wg-ddns
```

### Linux

```shell
wg-ip -i YOUR_INTERFACE -t 180
```
