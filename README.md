# SynchroDB
![Leonardo_Phoenix_10_Depict_a_futuristic_hightech_illustration_3](https://github.com/user-attachments/assets/ef39414c-1423-49ad-9089-94ba4bb8b7cc)

My attempt at making a distributed KV store

> [!IMPORTANT]
> <strong>⚠️ The project is still in early development, expect bugs, safety issues, and things that don't work ⚠️</strong> 

# Help

### How to generate certificates

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout server-key.pem -out server-cert.pem
```


### Test using openssl (not recommended)

```
openssl s_client -connect localhost:8000
```


### Test using client

```
go run cmd/client/main.go
```
or build and run the binary for better performance during benchmarking (see --help for the client).
<br>
The client will automatically authenticate if the server config is in the default path else one needs to provide the path to the server config file.

### Benchmark results

> [!IMPORTANT]
> <strong>⚠️ Take these results with a grain of salt as they are done locally on the same machine. This is just to get an idea of the database performance. ⚠️</strong> 

```
./synchrodb_client -benchmark -clients 100 -iterations 10000
+---------+----------+----------+----------+----------+----------------------+
| COMMAND | MIN (MS) | MAX (MS) | AVG (MS) | P99 (MS) | THROUGHPUT (OPS/SEC) |
+---------+----------+----------+----------+----------+----------------------+
| INCR    |    0.013 |   22.207 |    0.290 |    1.554 |           333261.512 |
| DECR    |    0.013 |   18.603 |    0.290 |    1.561 |           333261.512 |
| PING    |    0.012 |   19.856 |    0.289 |    1.539 |           333261.512 |
| SET     |    0.013 |   20.801 |    0.291 |    1.569 |           333261.512 |
| GET     |    0.013 |   19.879 |    0.290 |    1.557 |           333261.512 |
+---------+----------+----------+----------+----------+----------------------+
Successful clients: 100/100
Iterations per client: 10000
Total commands executed: 5000000
Total duration: 15.00 seconds
```