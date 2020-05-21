package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Response struct {
	Msg   string  `json:"msg"`
	Code int `json:"code"`
	Data ResUserInfo `json:"data"`
}
type ResUserInfo struct {
	Name string `json:"name"`
	Title string `json:"title"`
}

type ResGift struct {
	Msg string `json:"msg"`
	Code int `json:"code"`
}
func main() {
	uid := getUidFormWeb()
	//uid := getUserList()
	//fmt.Println(uid)
	//return
	in := dhmIn()
	dhmList := []string{}
	if in != "" {
		dhmList =  append(dhmList, in)
	} else {
		dhmList = append(dhmList,
			"家族联赛暗部冠军","忍忍向前冲","忍忍欧气爆棚","gg666",
			"ut26nbww","ecd93tz5","小改改最美","大声说出520")
	}
	//uid = []string{
	//	"814743375","120985832","309958882","733495838","605861637",
	//	"808259191","815819975","104342408","509853564","505174348",
	//	"120197976","228189697","207312785","622375509","216274841",
	//	"215676537","712534118","100537184","120125384","814785023",
	//	"608715317","521890764","107770832","419911467","133048744",
	//	"605993053","604010397",
	//}
	var wg sync.WaitGroup
	for _,id := range uid{
		wg.Add(1)
		go getInfo(id,&wg,dhmList)
	}
	wg.Wait()
	fmt.Scanf("按任意键结束")
}
// 输入兑换码
func dhmIn() string {
	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("请输入兑换码:\n")
	//fmt.Printf("请输入兑换码(若不输入则使用原有兑换码):\n")
	input.Scan()
	line := input.Text()
	return line
}
// 通过网络获取uid
func getUidFormWeb() []string {
	resp, err := http.Get("https://docs.qq.com/dop-api/opendoc?normal=1&id=DWVlTaHlPWkpVTnFB")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	uidList := []string{}
	for scanner.Scan() {
		str := uni2str(scanner.Text())
		if len(str) == 9 {
			uidList = append(uidList,str)
			//fmt.Println(str)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return uidList
}
// 读取uid文件
func getUserList() []string {
	f, err := os.Open("uid.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	uidList := []string{}
	for {
		//遇到\n结束读取
		b, errR := buf.ReadBytes('\n')
		if errR != nil {
			if errR == io.EOF {
				break
			}
			fmt.Println(errR.Error())
		}
		sb := string(b)
		i := len(sb)
		uidList = append(uidList,sb[:i-1])
	}
	return uidList
}
// 获取用户信息
func getInfo(id string,wg *sync.WaitGroup,dhmList []string)  {
	resp, err := http.Get("https://statistics.pandadastudio.com/player/simpleInfo?uid="+ id)
	if err != nil {
		panic(err)
	}
	defer wg.Done()
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str := uni2str(scanner.Text())
		res := &Response{}
		json.Unmarshal([]byte(str), &res)
		if res.Code == 0 && res.Msg != "未找到玩家信息"{
			var wg1 sync.WaitGroup
			for _,dhm := range dhmList{
				wg1.Add(1)
				go getDhmInfo(id, res.Data.Name, dhm, &wg1)
			}
			wg1.Wait()
		}
		//fmt.Println(*res)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
//获取礼包
func getDhmInfo(id string,name string, dhm string, wg *sync.WaitGroup) bool {
	resp, err := http.Get("https://statistics.pandadastudio.com/player/giftCode?uid="+ id +"&code=" + dhm)
	if err != nil {
		panic(err)
	}
	defer wg.Done()
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		str := uni2str(scanner.Text())
		res := &ResGift{}
		json.Unmarshal([]byte(str), &res)
		if res.Code == 425{
			fmt.Println(name + " | 已领取过 | " +dhm)
		} else if res.Code == 417 {
			fmt.Println(dhm+" | 是不存在的兑换码")
			return false
		} else if res.Code == 424 {
			fmt.Println(dhm+" 是已过期兑换码")
			return false
		} else {
			fmt.Println(name + " | 领取 " + dhm +" | 礼包成功！")
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return true
}
//转码中文
func uni2str(str string) string {
	buf := bytes.NewBuffer(nil)
	i, j := 0, len(str)
	for i < j {
		x := i + 6
		if x > j {
			buf.WriteString(str[i:])
			break
		}
		if str[i] == '\\' && str[i+1] == 'u' {
			hex := str[i+2 : x]
			r, err := strconv.ParseUint(hex, 16, 64)
			if err == nil {
				buf.WriteRune(rune(r))
			} else {
				buf.WriteString(str[i:x])
			}
			i = x
		} else {
			buf.WriteByte(str[i])
			i++
		}
	}
	return buf.String()
}