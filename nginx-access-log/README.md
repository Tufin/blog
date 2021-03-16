# Sending Nginx Access Logs to a Golang Server in UDP

Implementation of [this article](TBD)

#### Running nginx using docker
```
$ docker run -it --rm --name my-nginx -v $PWD/nginx.conf:/etc/nginx/nginx.conf -p 8085:8085 nginx:1.19-alpine
```