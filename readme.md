GPU拓扑:
```
      GPU0  GPU1  GPU2  GPU3  GPU4  GPU5  GPU6  GPU7
GPU0   X    PIX   PHB   PHB   SOC   SOC   SOC   SOC
GPU1  PIX    X    PHB   PHB   SOC   SOC   SOC   SOC
GPU2  PHB   PHB    X    PIX   SOC   SOC   SOC   SOC
GPU3  PHB   PHB   PIX    X    SOC   SOC   SOC   SOC
GPU4  SOC   SOC   SOC   SOC    X    PIX   PHB   PHB
GPU5  SOC   SOC   SOC   SOC   PIX    X    PHB   PHB
GPU6  SOC   SOC   SOC   SOC   PHB   PHB    X    PIX
GPU7  SOC   SOC   SOC   SOC   PHB   PHB   PIX    X
```



GPU使用详情:
```
# docker inspect -f "{{.HostConfig.Devices}}" d431d5d521de
[{/dev/nvidia1 /dev/nvidia1 mrw} {/dev/nvidia3 /dev/nvidia3 mrw} {/dev/nvidiactl /dev/nvidiactl mrw} {/dev/nvidia-uvm /dev/nvidia-uvm mrw} {/dev/nvidia-uvm-tools /dev/nvidia-uvm-tools mrw}]

# docker inspect -f "{{.HostConfig.Devices}}" f6ba22c35a76
[{/dev/nvidia0 /dev/nvidia0 mrw} {/dev/nvidiactl /dev/nvidiactl mrw} {/dev/nvidia-uvm /dev/nvidia-uvm mrw} {/dev/nvidia-uvm-tools /dev/nvidia-uvm-tools mrw}]
```
由此可以看出当前宿主机上的两个容器d431d5d521de和f6ba22c35a76 分别占用了GPU1和GPU3(GPU1和GPU3是通过PHB通信的)以及GPU0.

查看宿主机当前各个容器占用的GPU卡信息:
```
# go run nvidia-dockergetgpus.go
Container and GPU devices:
{
  "ContainerName": "k8s_gpu_gpu-2053395734-hd2rn_pe_b2d5be09-e6f0-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia5,/dev/nvidia6",
  "Nums": "2"
}
{
  "ContainerName": "reverent_gates",
  "DeviceNum": "/dev/nvidia2",
  "Nums": "1"
}
{
  "ContainerName": "k8s_dockertest5_dockertest5-3672797148-rmnfd_pe_d2001077-e6e5-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia4",
  "Nums": "1"
}
{
  "ContainerName": "k8s_gpu_gpu-2053395734-429tz_pe_3623364d-e553-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia1,/dev/nvidia3",
  "Nums": "2"
}
{
  "ContainerName": "k8s_test_test-4227486117-zqnlj_pe_cd46560e-e48d-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia0",
  "Nums": "1"
}
```

根据当前分配状态动态分配GPU卡信息：
```
$ go run nvidia-dockergetgpus.go
Container and GPU devices:
{
  "ContainerName": "k8s_dockertest5_dockertest5-3672797148-rmnfd_pe_d2001077-e6e5-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia4",
  "Nums": "1"
}
{
  "ContainerName": "k8s_gpu_gpu-2053395734-429tz_pe_3623364d-e553-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia1,/dev/nvidia3",
  "Nums": "2"
}
{
  "ContainerName": "k8s_test_test-4227486117-zqnlj_pe_cd46560e-e48d-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia0",
  "Nums": "1"
}

$ go run getFreegpus.go 2 ,0,1,3,4
分配的GPU个数:2 已分配的GPU卡信息:,0,1,3,4
gpu nums: 8
分配到的GPU信息如下: [5 6]
,5,6
```
根据获取的GPU卡信息进行分配GPU容器:
```
$  NV_GPU=,5,6 nvidia-docker run -itd --name  idockerhub.jd.com/nvidia-docker/cuda8.0-runtime:centos6-17-10-19 bash

$ ./getgpus
 {
  "ContainerName": "biaoge",
  "DeviceNum": "/dev/nvidia5,/dev/nvidia6",
  "Nums": "2"
}
```


那么此时我们在根据第一个容器的规格扩容一个同规格(2颗GPU卡)的GPU容器
```
# docker inspect -f "{{.HostConfig.Devices}}" df70ac8e58d1
[{/dev/nvidia7 /dev/nvidia7 mrw} {/dev/nvidia2 /dev/nvidia2 mrw} {/dev/nvidiactl /dev/nvidiactl mrw} {/dev/nvidia-uvm /dev/nvidia-uvm mrw} {/dev/nvidia-uvm-tools /dev/nvidia-uvm-tools mrw}]
```
问题来了,当前新扩容的容器占用的GPU卡设备竟然是GPU7和GPU2,那么根据上述的GPU拓扑图来看，当前这两块GPU卡应该是跨CPU核心的.GPU2和GPU7是通过SOC通信的


`注意:`默认情况下，只要GPU容器正常停止，就可以认为GPU卡设备已经被卸载，其上的GPU卡就可以被其他容器使用。因此在检测容器的GPU卡占用情况只需要去检测`UP`状态的容器

```
# 查看一个已经退出的GPU容器 可以发现使用的是GPU5 和GPU6设备
$ docker inspect -f "{{.HostConfig.Devices}}" prickly_bartik
[{/dev/nvidiactl /dev/nvidiactl rwm} {/dev/nvidia-uvm /dev/nvidia-uvm rwm} {/dev/nvidia-uvm-tools /dev/nvidia-uvm-tools rwm} {/dev/nvidia5 /dev/nvidia5 rwm} {/dev/nvidia6 /dev/nvidia6 rwm}]

# 使用k8s调度一个两个卡的容器，发现依然可以使用这两块设备，因为上面那个容器挂了，可以被重新使用
{
  "ContainerName": "k8s_gpu_gpu-2053395734-hd2rn_pe_b2d5be09-e6f0-11e7-9beb-ecf4bbc19ea8_0",
  "DeviceNum": "/dev/nvidia5,/dev/nvidia6",
  "Nums": "2"
}

```
