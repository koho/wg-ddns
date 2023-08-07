# wg-svcb
WireGuard SVCB Endpoint

## Usage

Install the binary using the following command:

```shell
go install github.com/koho/wg-svcb@latest
```

### Windows

Install the background service:

```shell
sc create wg-svcb binPath= "C:\Users\Administrator\go\bin\wg-svcb.exe -i YOUR_INTERFACE -t 180" DisplayName= "WireGuard SVCB" start= auto
```

Start the service:

```shell
net start wg-svcb
```
