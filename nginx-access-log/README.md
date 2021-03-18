# Sending Nginx Access Logs to a Golang Server in UDP

Implementation of [this article](https://tufin.medium.com/sending-nginx-access-logs-to-a-golang-server-over-syslog-98b9108f40e1)

#### Running nginx using docker
```
$ docker run -it --rm --name my-nginx -v $PWD/nginx.conf:/etc/nginx/nginx.conf -p 8085:8085 nginx:1.19-alpine
```
