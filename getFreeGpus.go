/**
 * @File Name: getFreeGpus.go
 * @Author:
 * @Email:
 * @Create Date: 2017-12-23 14:12:32
 * @Last Modified: 2017-12-23 14:12:06
 * @Description:根据已分配的GPU卡信息以及需要分配的GPU个数来筛选可分配的GPU卡信息
 */
package main
import (
	"os"
	"fmt"
	"os/exec"
	"strconv"
	"sort"
	"strings"
)

var (
	alloc_gpus,midn int   //定义分配的gpus个数以及GPU列表的中值以及中值在可用列表的下标
	midv,alloc_gpu	string
	gpus_pool,freeGpus,gpus []string //定义宿主机GPU卡信息以及可用的GPU卡列表以及最终分配的gpus列表
)

func main() {
	args := os.Args
	if len(args) != 3 {
		fmt.Printf("Usage: %s alloc_cnt used_gpu_ids\n",args[0])
		os.Exit(1)
	}

	/*
	alloc_cnt,err := strconv.Atoi(args[1])
	if err != nil { fmt.Printf("please check the alloc_cnt ",err.Error()) }
	*/

	if alloc_cnt,err := strconv.Atoi(args[1]); err == nil {
		alloc_gpus = alloc_cnt
	}

	used_gpus_id := args[2]

	fmt.Printf("分配的GPU个数:%v 已分配的GPU卡信息:%v\n",alloc_gpus,used_gpus_id)

	//物理机的gpu个数
	gpus := getGpus()

	fmt.Println("gpu nums:",gpus)
	freeGpu := GetFreeGpus(alloc_gpus,used_gpus_id)
	for i := 0;i<len(freeGpu);i++ {
		alloc_gpu = alloc_gpu+","+freeGpu[i]
	}

	fmt.Println(alloc_gpu)

}

func GetFreeGpus(alloc int,usedGpus string) (FreeGpus []string){
	//获取宿主机的GPU个数并构造GPU卡列表
	//gpu_pool = [0 1 2 3 4 5 6 7]
	allGpus := getGpus()
	for i := 0;i < allGpus;i++ {
		gpus_pool = append(gpus_pool,strconv.Itoa(i))
	}

	//对已使用的GPU卡信息进行排序并构造已使用列表
	used := strings.Split(usedGpus,",")
	sort.Strings(used)

	//根据gpu卡列表和已使用列表构造可用gpu列表并计算可用列表中的中值
        if isslice,diff := checkSlice(used,gpus_pool); isslice == false {
		//构造可用的GPU卡列表
		freeGpus = diff
		//计算gpu卡列表中的中值在可用gpu卡列表中的位置
		for n,diffmid := range freeGpus {
			if num,err := strconv.Atoi(diffmid);err == nil {
				if num < allGpus/2 {
					midn = n
					midv = diffmid
				}
			}
		}
	}

	freeGpuNums := len(freeGpus)

	//全部GPU[0 1 2 3 4 5 6 7],中值是3，该值在可用GPU列表的位置为midn
	//如果在可用列表的前midn个元素中可以满足分配的卡，就从前半段分。如果前半段不够就判断后半段是否有足够的卡来分给用户，如果也不够，则退出，不能自动分配到近亲缘性的卡，如果够直接分配
	if alloc > freeGpuNums {
		fmt.Println("无可分配的GPU资源")
		os.Exit(1)
	}

	//当有8卡的时候,midn相当于gpu3在可用列表里的位置
	//当midn不等于0时，说明第一组GPU卡没有分配完
	//fmt.Println(midn)
	if midn > 0 {
	    	if alloc <= midn+1 {
			gpus = freeGpus[0:alloc_gpus]
		} else if len(freeGpus[midn+1:]) >= alloc {
			gpus = freeGpus[midn+1:midn+1+alloc_gpus]
		} else {
			fmt.Println("没有亲缘性的GPU节点,请检查!如有需要请手动指定多跨CPU的卡设备")
			os.Exit(2)
	}
	//midn = 0
	} else if midn = 0;freeGpus[0] == midv {
		if alloc != 1 {
			gpus = freeGpus[midn+1:midn+1+alloc_gpus]
		}else { gpus = freeGpus[midn:midn+alloc_gpus] }
	} else {
		gpus = freeGpus[midn:midn+alloc_gpus]
	}


	fmt.Println("分配到的GPU信息如下:",gpus)
	return gpus
}



//获取物理机的GPU个数
func getGpus() (gpu int) {

	gpus,err := exec.Command("/bin/bash","-c",`nvidia-smi -L | wc -l`).Output()
	if err != nil {
		fmt.Println("Failed to get the gpu info with nvidia-smi command",err.Error())
		os.Exit(1)
	}
	gpunums,err := strconv.Atoi(strings.Replace(string(gpus),"\n","",-1))
	if err != nil { os.Exit(1) }
	return gpunums


}

//对比已使用GPU卡以及全部GPU卡列表，计算出可分配的GPU列表
func checkSlice(a,b []string) (isIn bool,c []string) {
	for _,valueOfb := range b{
		temp := valueOfb
		for j := 0;j < len(a); j++ {
			if temp == a[j] {
				break
			} else {
				if len(a) == (j+1) {
					c = append(c,temp)
				}
			}

		}
	}
	if len(c) == 0 {
		isIn = true
	} else {
		isIn = false
	}
	return isIn,c
}
