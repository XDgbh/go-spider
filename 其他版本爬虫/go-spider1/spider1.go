package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"	//正则表达式
	"strconv"
	"time"
)

//定义新的数据类型
type Spider struct {
	url    string
	header map[string]string
}

//定义 Spider get的方法，获取到html文本
func (keyword Spider) get_html_header() string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", keyword.url, nil)
	if err != nil {
	}
	for key, value := range keyword.header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	}
	return string(body)

}
func parse() {
	header := map[string]string{
		"Host":                      "movie.douban.com",
		"Connection":                "keep-alive",
		"Cache-Control":             "max-age=0",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.143 Safari/537.36",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Referer":                   "https://movie.douban.com/top250",
	}

	//在当前目录下，创建excel文件
	f, err := os.Create("./豆瓣电影TOP250.xlsx")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	//写入标题，\t换到下一个制表符
	f.WriteString("电影名称" + "\t" + "评分" + "\t" + "评价人数" + "\t" + "\r\n")

	//循环每页解析并把结果写入excel，i表示第i页，总共0-9页
	for i := 0; i < 10; i++ {
		fmt.Println("\n正在抓取第" + strconv.Itoa(i) + "页........................")
		url := "https://movie.douban.com/top250?start=" + strconv.Itoa(i*25) + "&filter="
		spider := &Spider{url, header}
		html := spider.get_html_header()

		//评价人数
		pattern2 := `<span>(.*?)评价</span>` //匹配正则表达式如HTML中： <span>881457人评价</span>
		rp2 := regexp.MustCompile(pattern2)	//编译解析正则表达式，
		find_txt2 := rp2.FindAllStringSubmatch(html, -1)	//返回所有n个匹配的内容,同时返回子表达式匹配的内容
		// {完整匹配项, 子匹配项, 子匹配项, ...},  //子匹配项就是括号()里面的,返回的内容存入n*1维字符串slice中
		//find_txt2==["<span>881457人评价</span>","881457人"]，也就是说find_txt2[1]=="881457人"
		//评分
		pattern3 := `property="v:average">(.*?)</span>`	//匹配正则表达式如HTML中： property="v:average">9.6</span>
		rp3 := regexp.MustCompile(pattern3)
		find_txt3 := rp3.FindAllStringSubmatch(html, -1)

		//电影名称
		//pattern4:=`class="title">(.*?)</span>`	//这个就不行了，因为有两行匹配到这个字符串的<span class="title">肖申克的救赎</span>
		//<span class="title">&nbsp;/&nbsp;The Shawshank Redemption</span>
		pattern4 := `img alt="(.*?)" src=`	//匹配正则表达式如HTML中：img alt="肖申克的救赎" src=
		rp4 := regexp.MustCompile(pattern4)
		find_txt4 := rp4.FindAllStringSubmatch(html, -1)

		// 写入UTF-8 BOM
		//f.WriteString("\xEF\xBB\xBF")
		//  打印全部数据和写入excel文件,j表示的是这一个页面中的第j个电影
		//find_txt4[j][1]，这个1用的好，0对应的字符串是完全匹配的字符串，1对应的是括号中子表达式匹配的字符串
		for j := 0; j < len(find_txt2); j++ {
			fmt.Printf("%s \t%s \t%s\n", find_txt4[j][1], find_txt3[j][1], find_txt2[j][1])
			f.WriteString(find_txt4[j][1] + "\t" + find_txt3[j][1] + "\t" + find_txt2[j][1] + "\t" + "\r\n")

		}
	}
}

func main() {
	t1 := time.Now() // get current time
	parse()
	elapsed := time.Since(t1)
	fmt.Println("爬虫结束,总共耗时: ", elapsed)

}
