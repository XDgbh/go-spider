//go语言简单实现爬虫
/*
实现爬虫的基本思想是：
	首先定义一个入口页面，然后一个页面会有其他页面的URL，于是从当前页面获取到这些URL加入到爬虫的抓取队列中，
	然后进入到新页面后再递归的进行上述的操作，其实说来就跟深度遍历或广度遍历一样。
*/

/*
具体操作步骤：
1、
2、
3、
4、
*/

package main

//完全不借助第三方的框架,通过go sdk的标准库来实现一个爬虫应用,主要用到的包如下
import (
	"encoding/xml" //encoding/xml 解析xml的包
	"fmt"
	"io/ioutil" //io/ioutil io处理的工具包
	"log"
	"math/rand"
	"net/http" //net/http 标准库里内建了对http协议的支持,实现了一个http client,可以直接通过其进行get,post等请求
	"regexp"   //regexp go sdk中的正则表达式包
	"runtime"
	_ "runtime"
	"strings" //strings 不像java的String是一个引用类型,go语言中的字符串类型是一个内建的基础类型, 而且go语言默认只支持UTF-8编码,strings包实现了一些简单的针对utf-8字符串操作的函数
	"time"
)

var urlChannel = make(chan string, 200) //首先声明一个全局的urlchannel用来同步开启的多个routines在某个页面获取的<a> 标签的href属性,chan中存入string类型的href属性,缓冲200
//<a> 标签的 href 属性用于指定超链接目标的URL（如：<a href="#/URL">name</a>）。如果用户选择了<a>标签中的内容，那么浏览器会尝试检索并显示href 属性指定的URL所表示的文档，或者执行JavaScript表达式、方法和函数的列表。

var atagRegExp = regexp.MustCompile(`<a[^>]+[(href)|(HREF)]\s*\t*\n*=\s*\t*\n*[(".+")|('.+')][^>]*>[^<]*</a>`) //声明在html文档中获取<a> 的正则表达式,
//以Must前缀的方法或函数都是必须保证一定能执行成功的,否则将引发一次panic

//当进入main函数时,将启动一个goroutine来从入口url=”http:/blog.csdn.net”开始爬取(Spy函数)页面内容分析<a> 标签
//接下来通过for range urlChannel来循环取出爬取到的<a> 标签中的href属性,并再次开启一个新的goroutine来爬取这个href属性对应的html文档内容
func main() {
	go Spy("http://blog.csdn.net")
	//go Spy("http://www.iteye.com/")
	for url := range urlChannel {
		fmt.Println("routines num = ", runtime.NumGoroutine(), "chan len = ", len(urlChannel)) //通过runtime可以获取当前运行时的一些相关参数等
		go Spy(url)
	}
	fmt.Println("a")
}

/*函数功能：解析<a> 元素
获取<a>标签，并用xml库解析处理这个标签，返回我们需要的href值和content内容
<a> 可以当做一份xml文档(只有一个a为根节点的简单xml)来解析出href/HREF属性,通过go标准库中xml.NewDecoder来完成
*/
func GetHref(atag string) (href, content string) {
	inputReader := strings.NewReader(atag)
	decoder := xml.NewDecoder(inputReader)
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			for _, attr := range token.Attr {
				attrName := attr.Name.Local
				attrValue := attr.Value
				if strings.EqualFold(attrName, "href") || strings.EqualFold(attrName, "HREF") {
					href = attrValue
				}
			}
		// 处理元素结束（标签）
		case xml.EndElement:
		// 处理字符数据（这里就是元素的文本）
		case xml.CharData:
			content = string([]byte(token))
		default:
			href = ""
			content = ""
		}
	}
	return href, content
}

/*
函数功能：
1、由于每个爬取goroutine都是调用Spy函数来分析一个url对应的html文档,所以需要在函数开始就defer 一个匿名函数来处理(recover)可能出现的异常(panic),
	防止异常导致程序终止,defer执行的函数会在当前函数执行完成后结果返回前执行,无论该函数是panic的还是正常执行.
2、由于go内建了对http协议的支持,可以直接通过http包下的http.Get或则http.Post函数来请求url.
但由于大部分网站对请求都有防范DDOS等的限制,需要自定义请求的header,设置代理服务器(CSDN好像对同一IP的请求平率限制并不严格,iteye亲测很严格,每分钟上万会被封住IP)等操作,
3、可以使用http包下的http.NewRequest(method, urlStr string, body io.Reader) (*Request, error)函数,
4、然后通过Request的Header对象设置User-Agent,Host等,
5、最后调用http包下内置的DefaultClient对象的Do方法完成请求.
6、返回响应成功的状态码200,即拿到服务器响应后(*Response)通过ioutil包下的工具函数转换为string,找出文档中的<a>标签 分析出href属性,存入urlChannel中.
*/
func Spy(url string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[E]", r)
		}
	}()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", GetRandomUserAgent())
	client := http.DefaultClient
	res, e := client.Do(req)
	if e != nil {
		fmt.Errorf("Get请求%s返回错误:%s", url, e)
		return
	}

	if res.StatusCode == 200 {
		body := res.Body
		defer body.Close()
		bodyByte, _ := ioutil.ReadAll(body)
		resStr := string(bodyByte)
		atag := atagRegExp.FindAllString(resStr, -1)
		for _, a := range atag {
			href, _ := GetHref(a) //调用获取href的函数

			if strings.Contains(href, "article/details/") { //Contains()函数返回bool值，判断后一个字符串是不是前一个字符串的子串
				fmt.Println("☆", href)
			} else {
				fmt.Println("□", href)
			}
			urlChannel <- href
		}
	}
}

//随机伪造User-Agent
var userAgent = [...]string{"Mozilla/5.0 (compatible, MSIE 10.0, Windows NT, DigExt)",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
	"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
	"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0,",
	"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
	"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1"}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func GetRandomUserAgent() string {
	return userAgent[r.Intn(len(userAgent))]
}
