/**
 * @File Name: getUsedGpuinfo.go
 * @Author:
 * @Email:
 * @Create Date: 2017-12-23 14:12:55
 * @Last Modified: 2017-12-23 14:12:57
 * @Description: 获取GPU主机上各个容器分配的GPU卡信息
 */
package main
import (
	"os"
	"regexp"
	"os/exec"
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
)


type GpuInfo struct {
    ContainerName   string `json:container`
    DeviceNum       string `json:devicenum`
    Nums            string `json:nums`

}


func main() {
	fmt.Println("Container and GPU devices:")
	err := GetUsedGpu()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func GetUsedGpu() error {
    var AllUsedDevice []interface{}
    data := GpuInfo{}
    //UsedDeviceInfo := [](data)
    //[]byte
    GetConName,err := exec.Command("/bin/bash","-c",`docker ps --format="{{.Names}}"`).Output()
    if err != nil {
	fmt.Println("no container")
	os.Exit(1)
    }
    GetList := strings.Split(string(GetConName),"\n")
    ConNameList := GetList[0:len(GetList)-1]
    for _,name := range ConNameList {
        //获取容器的挂载设备返回带换行符的[]byte
        GetDeivceOfName,err := exec.Command("/bin/bash","-c",`docker inspect -f "{{.HostConfig.Devices}}" `+name).Output()
        if err != nil {
             fmt.Println("no get container device")
	     os.Exit(3)
        }
        //GetDeivceOfName [{/dev/nvidiactl /dev/nvidiactl rwm} {/dev/nvidia-uvm /dev/nvidia-uvm rwm} {/dev/nvidia-uvm-tools /dev/nvidia-uvm-tools rwm} {/dev/nvidia5 /dev/nvidia5 rwm} {/dev/nvidia6 /dev/nvidia6 rwm}]
        devicestrings := strings.Replace(string(GetDeivceOfName),"\n","",-1)
	//filter the cpu container
	if devicestrings != "[]" {
		var UsedDevice string
	        if re := regexp.MustCompile("k8s*"); re.MatchString(name) == true {
			UsedDevice = GetK8sGpuDevice(name,devicestrings)
		} else {
			UsedDevice = GetDockerGpuDevice(name,devicestrings)
		}
	        data.ContainerName = name
		data.DeviceNum = strings.Replace(UsedDevice,",","",1)
		data.Nums = strconv.Itoa(len(strings.Split(UsedDevice,","))-1)
		AllUsedDevice =  append(AllUsedDevice,data)
		/*
		result,err := json.Marshal(data)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(result))
		*/
	}

    }
    for _,InfoPerCon := range AllUsedDevice {
		//if data,err := json.Marshal(InfoPerCon); err == nil {
		if data,err := json.MarshalIndent(InfoPerCon,"","  "); err == nil {
			fmt.Println(string(data))
		}
    }
    return nil

}


//return ,/dev/nvidia0,/dev/nvidia1
func GetDockerGpuDevice(name,devices string)  (useddevice string){
    var gpus []string
    devicelist := strings.Split(devices," ")	//[]string
    gpulist := devicelist[9:len(devicelist)]
        for i := 1; i < len(gpulist); i  = i+3  {
		gpus = append(gpus,gpulist[i])
        }

    var alldevice string
    for i := 0;i < len(gpus);i = i+1  {
	alldevice = alldevice+","+gpus[i]

    }

    //fmt.Println(alldevice)
    return alldevice
}

func GetK8sGpuDevice(name,devices string)  (useddevice string){
    var gpus []string
    devicelist := strings.Split(devices," ")
    gpulist := devicelist[0:len(devicelist)-9]
        for i := 1; i < len(gpulist); i  = i+3  {
    		gpus = append(gpus,gpulist[i])
        }

    var alldevice string
    for i := 0;i < len(gpus);i = i+1  {
  	alldevice = alldevice+","+gpus[i]

    }

    //fmt.Println(alldevice)
    return alldevice
}
