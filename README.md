# Humpback Agent Intro

Get images
---
> `GET` - http://localhost:8500/dockerapi/v2/images
```javascript	
/*Response*/
{
  "Id": "dce39deae261fe3b6b67c3fbd185348ee87c714c505baefecadb407788bf2c50",
  "ParentId": "02c7a1039a3d5c2202a4db8a3cd9c09e3f7ee5513f18bb75f1783159979c3c3b",
	"RepoTags": [
    "dockerapi:latest"
  ],
  "Created": 1438758840,
  "VirtualSize": 353337635,
  "Labels": null
}
```

Get image detail info
---
> `GET` - http://localhost:8500/dockerapi/v2/images/`{imageID}`
```javascript
/*Response*/
{
  "Id": "b56964f6ef63d39d50345d23ed808797f34e4007a188e71b9d397b4f2427aa77",
  ...
  "Parent": "aa7fdf353ca4477f40c269c587eb18b2a1cfbda9f5e7de4dbcc69a207099a581",
  "Created": "2015-08-06T00:39:00.284394047Z",
  ...
  "Config": {
    "Hostname": "147d80e3f549",
    "ExposedPorts": {
      "22/tcp": {},
      "5800/tcp": {}
    },
    "Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],
    "Cmd": ["./dockerapi"],
    "Image": "aa7fdf353ca4477f40c269c587eb18b2a1cfbda9f5e7de4dbcc69a207099a581",
    "Entrypoint": null
  },
  "Architecture": "amd64",
  "VirtualSize": 353337635,
  ...
}
```

Pull image from registry
---
> `POST` - http://localhost:8500/dockerapi/v2/images
```javascript	
/*Request*/
{
  "Image": "dockerapi:2.2.0"
}
```

Delete image from server
---
> `DELETE` - http://localhost:8500/dockerapi/v2/images/`{imageID}`
		
Get containers
---
> `GET` - http://localhost:8500/dockerapi/v2/containers?`all=[true|false]`  
> `true` 获取所有container, `false` 表示获取正在运行的container, 默认为`false`
```javascript	
/*Response*/
[
  {
    "Id": "c8a78bc0c27400d56aa63a80ecba7644c3313242dd4565eee98ae190defb69b6",
    "Names": [
      "/dockerapi"
    ],
    "Image": "dockerapi:2.2.0",
    "ImageID": "",
    "Command": "./docker-api",
    "Created": 1470820025,
    "Ports": [],
    "Labels": {},
    "State": "",
    "Status": "Up 15 hours",
    "HostConfig": {
      "NetworkMode": "host"
    },
    "NetworkSettings": null,
    "Mounts": null
  }
]
```

Get container detail info
---
> `GET` - http://localhost:8500/dockerapi/v2/containers/`{containerID}`
```javascript	
/*Response*/
{
  "Id": "c8a78bc0c27400d56aa63a80ecba7644c3313242dd4565eee98ae190defb69b6",
  "Image": "dockerapi:2.2.0",
  "Command": "./docker-api",
  "Name": "dockerapi",
  "Ports": [
    {
      "PrivatePort": 8500,
      "PublicPort": 0,
      "Type": "tcp",
      "Ip": "0.0.0.0"
    }
  ],
  "Volumes": [
    {
      "ContainerVolume": "/var/run",
      "HostVolume": "/var/run"
    }
  ],
  "Env": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
  ],
  "HostName": "",
  "NetworkMode": "host",
  "status": {
    "Status": "",
    "Running": true,
    "Paused": false,
    "Restarting": false,
    "OOMKilled": false,
    "Dead": false,
    "Pid": 18928,
    "ExitCode": 0,
    "Error": "",
    "StartedAt": "2016-08-10T09:07:05.7151068Z",
    "FinishedAt": "0001-01-01T00:00:00Z"
  },
  "RestartPolicy": "no",
  "Extrahosts": null,
  "Links": null
  "CPUShares": 0,
  "Memory": 0,
  "SHMSize": 0
}
```
为了与创建container时传递的 `Request Body`一致，默认返回经过处理后的数据。如果想要docker返回的原始数据，请加上queryString **`originaldata=true`**

Create Container
---
> `POST` - http://localhost:8500/dockerapi/v2/containers
```javascript
/*Request*/
{
  "Name": "eggkeeper",
  "Command": "",          #为空则便是使用镜像内部默认的Cmd
  "Image":"eggkeeper:1.2.1",
  "Ports":[
    {
      "PrivatePort":8484,
      "PublicPort":8484,
      "Type":"tcp",       # tcp | udp
      "Ip":"10.16.75.23"  # 不指定ip则绑定到宿主机的所有ip
    },
    {
      "PrivatePort":22,
      "PublicPort":0,     # 0 表示自动分配宿主机端口 
      "Type":"tcp"		
    }
  ],
  "Volumes":[
    {
      "ContainerVolume":"/app-conf",
      "HostVolume":"/opt/app/app-conf"
    }
  ],
  "Env":[
    "env=gdev"
  ],
  "HostName":"testhostname",  # 仅当 networkmode 为 `bridge` 生效
  "NetworkMode":"bridge",     # host | bridge
  "RestartPolicy":"no",       # always | on-failure | no(default)
  "RestartRetryCount":"2",    # 仅当 restartpolicy 是 on-failure 时才有用
  "Extrahosts": [
    "hostname:IP"
  ],                          # 仅当 networkmode 为 `bridge` 生效
  "CPUShares": 500,
  "Memory": 200,              # MB      
  "Links": [                  # 使用规范详见docker官方文档
    "container_name:alias"
  ]    
}
```
```javascript	
/*Response*/
{
  "Id": "1f39cc7c9ac1da65a3e1d8b702f2b7ba118721d4a684bc3f502dfff224c79d03",
  "Name": "eggkeeper",
  "Warnings": null
}
```

Start/Stop... container
---
> `PUT` - http://localhost:8500/dockerapi/v2/containers
```javascript	
/*Request*/
{
  "Action": "[start|stop|restart|kill|pause|unpause]",
  "Container": "eggkeeper"
}	
```

Upgrade container's image
> `PUT` - http://localhost:8500/dockerapi/v2/containers
```javascript
/*Request*/
{
  "Action":"upgrade",
  "Container":"eggkeeper",
  "ImageTag":"1.2.1"
}
```
```javascript
/*Response*/
{
  "Id": "c4b33475ab59ecc1c4657447ea29bd5ec82e2c7d0aff70296ba4f86d13eaed55"  # new container's id
}
```

Rename container
---
> `PUT` - http://localhost:8500/dockerapi/v2/containers
```javascript
/*Request*/
{
  "Action":"rename",
  "Container":"eggkeeper",
  "NewName":"eggkeeper2"
}
```

Delete container
---
> `DELETE` - http://localhost:8500/dockerapi/v2/containers/`{containerID}`


# Other
执行 `glide install` 安装相关依赖包
