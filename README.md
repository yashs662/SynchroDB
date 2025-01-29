# SynchroDB
![Leonardo_Phoenix_10_Depict_a_futuristic_hightech_illustration_3](https://github.com/user-attachments/assets/ef39414c-1423-49ad-9089-94ba4bb8b7cc)

My attempt at making a distributed KV store

```
⚠️⚠️ This project is currently a work in progress and is not yet completed. ⚠️⚠️
```

# Help

### How to generate certificates

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server-key.pem -out server-cert.pem
```


### Test using openssl

```
openssl s_client -connect localhost:6379
```


### Test using client

```
go run cmd/client/main.go
```
or build and run the binary for better performance during benchmarking (see --help for the client).
<br>
The client will automatically authenticate if the server config is in the default path else one needs to provide the path to the server config file.
