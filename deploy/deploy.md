```shell

docker run -d \
--name=humpback-agent \
--net=host \
--restart=always \
--privileged \
-v /etc/localtime:/etc/localtime \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /var/lib/docker:/var/lib/docker \
-e HUMPBACK_AGENT_API_BIND=0.0.0.0:8018 \
-e HUMPBACK_SERVER_HOST={server-address}:{server-backend-port} \
-e HUMPBACK_VOLUMES_ROOT_DIRECTORY=/var/lib/docker \
docker.io/humpbacks/humpback-agent:develop
```