```shell

docker run -d \
--name=humpback-agent \
--net=host \
--restart=always \
--privileged \
-v /etc/localtime:/etc/localtime \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /var/lib/docker:/var/lib/docker \
-e PORT=8018 \
-e SERVER_ADDRESS={server-address}:{server-backend-port} \
docker.io/humpbacks/humpback-agent:develop
```