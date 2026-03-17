# Local Development Compose

### 添加端口转发

```pwsh
netsh interface portproxy add v4tov4 listenport=8001 listenaddress=0.0.0.0 connectport=8001 connectaddress=(wsl hostname -I)
```
