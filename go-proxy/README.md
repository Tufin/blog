# HTTP Client using proxy with Go
There are 3 options to "tell" a _Go_ client to use a proxy:

a. Set the `HTTP_PROXY` environment variable:
```bash
$ export HTTP_PROXY="http://ProxyIP:ProxyPort"
```
`HTTP_PROXY` environment variable will be used as the proxy URL for HTTP requests and HTTPS requests,
unless overridden by HTTPSProxy or NoProxy
    
b. Creating an HTTP client in _Go_ that MUST use a proxy:
```
proxy, _ := url.Parse("http://ProxyIP:ProxyPort")
httpClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
```
For more robust _HTTP Client_ [checkout](https://github.com/tufin/blog/go-proxy/common/http.go)

c. Change the default transport used by _Go_ "net/http" package. 
This would affect the entire program (including the default HTTP client)
```
proxy, _ := url.Parse("http://ProxyIP:ProxyPort")
http.DefaultTransport := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
```
# Tor as a Proxy
[Tor](https://www.torproject.org/) aim to defend against tracking and surveillance.
You might want to write an application that would not be traced (like a crawler). 
An easy solution can be using _Tor_ as a proxy.

## Tor Installation
Few options for _Tor_ installation:
1. Installing [Tor browser](https://tb-manual.torproject.org/installation/)
or there are browsers like [brave](https://brave.com/) that comes with _Tor_ browse as an option
2. Install as a proxy service on your computer, see [Tor docs](https://2019.www.torproject.org/docs/tor-doc-osx.html.en)
3. Run _Tor_ inside a Docker container

### Running _Tor_ inside a Docker container
Running _Tor_ inside a Docker container makes it easy if you want to package your application with the _Tor_.
For example, if you want to run batch of HTTP calls as part of CI workflow.

#### How to?
1. Copy the follow into a [Dockerfile.tor](https://github.com/tufin/blog/go-proxy/Dockerfile.tor)
```
FROM alpine:edge
RUN apk update && apk add tor
EXPOSE 9150
CMD tor -f /etc/torrc
```
2. Create a docker image name _tor_ (optional)
```
docker build -t tor -f Dockerfile .
```
3. Run created image
```
docker run -d --rm --name tor -p 9150:9150 tor
```
Or use image from github (if you want to skip 2)
```
docker run -d --rm --name tor -p 9150:9150 tufin/tor
```
After that, you'll have a _Tor_ proxy running on `127.0.0.1:9150`
so go ahead and configure your browser to use a SOCKS proxy on `127.0.0.1:9150`,
or use _Tor_ as a proxy for _Go_ client

## Using _Tor_ as a Proxy for Go Client
Like we did above, just replace the URL to the running _Tor_:
```
proxy, _ := url.Parse("socks5://127.0.0.1:9050")
http.DefaultTransport := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
```
For more robust _HTTP Client_ [checkout](https://github.com/tufin/blog/go-proxy/common/http.go)

## Using _Tor_ as an egress proxy inside a _Kubernetes_ cluster
You might want that some of your services that part of a k8s will run HTTP calls using _Tor_.
In order to do that let's combine all above, and a little more :)

Our architecture will look like:

Service A --
            |--> Tor Service -> Cloud
Job B     --

### Deploy a _Tor_ Egress Proxy
Follow yaml contains k8s service, and deployment for _Tor_ (same docker image as above):
```
apiVersion: v1
kind: Service
metadata:
  name: tor
  namespace: default
  labels:
    app: tor
spec:
  selector:
    app: tor
  ports:
    - name: http
      port: 9050
      targetPort: 9050
      protocol: TCP
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tor
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: tor
    spec:
      containers:
        - name: tor
          image: tufin/tor
          imagePullPolicy: Always
          ports:
            - containerPort: 9050
```
Now, we let's configure a _Go_ service to use our _Tor_ Egress proxy by adding `HTTP_PROXY` header. 
`HTTP_PROXY` environment variable will be used as the proxy URL.
```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tor
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: tor
    spec:
      containers:
        - name: tor
          image: tufin/tor
          imagePullPolicy: Always
          ports:
            - containerPort: 9050
```