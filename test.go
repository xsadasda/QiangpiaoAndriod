package fuben

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/robfig/cron"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	urlIndex       = "https://wechartdemo.zckx.net/API/TicketHandler.ashx"
	urlOrder       = "https://wechartdemo.zckx.net/Ticket/SaveOrder?"
	userName       = "王硕"
	userPhone      = "19861109837"
	userIdentityNo = "370681200011276012"
	openId         = "oSkwT5SDTkZYOTGapBY6G3sJJWrA"

	//循环次数
	epoch = 2
)

type item struct {
	Num        int    `json:"num"`
	StyNo      string `json:"styNo"`
	StrId      string `json:"strId"`
	DaySurplus string `json:"daySurplus"`
	BeginH     string `json:"beginH"`
}

type Index struct {
	Hour  string `json:"hour"`
	Items []item `json:"items"`
}

type timeItem struct {
	MinDate  string `json:"minDate"`
	MaxDate  string `json:"maxDate"`
	Strategy string `json:"strategy"`
}

type styleItem struct {
	StyleNo       string `json:"styleNo"`
	StyleGroupNo  string `json:"styleGroupNo"`
	StyleName     string `json:"styleName"`
	Price         string `json:"price"`
	DiscountPrice string `json:"discountPrice"`
	TicketNum     int    `json:"ticketNum"`
	SolutionNo    string `json:"solutionNo"`
	ProjectNo     string `json:"projectNo"`
}
type userItem struct {
	UserName       string `json:"userName"`
	UserPhone      string `json:"userPhone"`
	UserIdentityNo string `json:"userIdentityNo"`
}

type orderBody struct {
	UserDate      string      `json:"userDate"`
	TimeList      []timeItem  `json:"timeList"`
	TotalPrice    int         `json:"totalprice"`
	StyleInfoList []styleItem `json:"styleInfoList"`
	UserInfoList  []userItem  `json:"userInfoList"`
	OpenId        string      `json:"openId"`
	SellerNo      string      `json:"sellerNo"`
}
type responseBody struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
	Data    string `json:"Data"`
}

var beginH = 0
var endH = 0
var date = ""

// 抢到票就变成1
var FLAG = 0

var styleListStr = []int64{1000001077, 1000001078, 1000001079, 1000001080, 1000001081, 1000001085, 1000001086, 1000001087, 1000001088, 1000001089, 1000001090, 1000001091, 1000001092, 1000001093, 1000001094, 1000001095, 1000001096, 1000001097}

func getIndex() []Index {

	var data = url.Values{}
	data.Set("dataType", "json")
	data.Set("date", date)
	data.Set("minHour", "6")
	data.Set("maxHour", "22")
	data.Set("projectNo", "1000000635")
	data.Set("openId", openId)
	data.Set("styleListStr", "1000001077,1000001078,1000001079,1000001080,1000001081,1000001085,1000001086,1000001087,1000001088,1000001089,1000001090,1000001091,1000001092,1000001093,1000001094,1000001095,1000001096,1000001097")
	data.Set("method", "GetDataList")

	reader := strings.NewReader(data.Encode())
	req, _ := http.NewRequest("POST", urlIndex, reader)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Referer", "https://wechartdemo.zckx.net/Ticket/SportHallsKO?&projectNo=1000000635&openId="+openId)
	req.Header.Add("Host", "wechartdemo.zckx.net")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	temp, err := io.ReadAll(resp.Body)

	index := make([]Index, 0)
	err = json.Unmarshal(temp, &index)
	if err != nil {
		log.Fatal(err)
	}
	return index

}

func getOoder(index []Index, beginH int, endH int) map[string]int {
	//strNo 时间信息 strId场次信息
	flag := 1
	for i := 5; i < 5+18; i++ { // 场地循环
		i = i % 18
		j := 0
		for j = beginH; j < endH; j++ { //时间循环
			if index[j-6].Items[i].Num == 0 {
				flag = 0
				break
			}
		}
		if flag == 1 {
			m := map[string]int{
				"time":    j - 6 - (endH - beginH),
				"session": i,
			}
			return m

		}
		flag = 1
	}
	return nil
}

func prepareOrder(index []Index, strId map[string]int) *http.Request {

	timeI := []timeItem{}
	for i := 0; i < endH-beginH; i++ {
		time := strId["time"] + i
		session := strId["session"]
		str2int, _ := strconv.Atoi(index[time].Items[session].BeginH)
		int2str := strconv.Itoa(str2int + 1)
		timeI = append(timeI, timeItem{
			MinDate: index[time].Items[session].BeginH + ":00",
			//todo maxDate 怎么妥善的处理
			MaxDate:  int2str + ":00",
			Strategy: index[time].Items[session].StrId,
		})
	}
	time := strId["time"]
	session := strId["session"]
	styleInfoList := []styleItem{}

	styleInfoList = append(styleInfoList, styleItem{
		StyleNo:       index[time].Items[session].StyNo,
		StyleGroupNo:  "",
		StyleName:     "",
		Price:         "",
		DiscountPrice: "",
		TicketNum:     1,
		SolutionNo:    "",
		ProjectNo:     "",
	})

	userInfoList := []userItem{}
	userInfoList = append(userInfoList, userItem{
		UserName:       userName,
		UserPhone:      userPhone,
		UserIdentityNo: userIdentityNo,
	})

	body := orderBody{
		UserDate:      date,
		TimeList:      timeI,
		TotalPrice:    0,
		StyleInfoList: styleInfoList,
		UserInfoList:  userInfoList,
		OpenId:        openId,
		SellerNo:      "weixin",
	}

	var data = url.Values{}
	postBody, _ := json.Marshal(body)
	data.Set("dataType", "json")
	data.Set("orderJson", string(postBody))
	reader := strings.NewReader(data.Encode())
	req, _ := http.NewRequest("POST", urlOrder, reader)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Referer", "https://wechartdemo.zckx.net/Ticket/SportHallsKO?&projectNo=1000000635&openId="+openId)
	req.Header.Add("Host", "wechartdemo.zckx.net")

	//sendOrderOrder(req)
	return req

}

func sendOrderOrder(request *http.Request) {
	client := &http.Client{Timeout: 3 * time.Second}
	println("开始抢票")
	resp, _ := client.Do(request)
	if resp != nil {
		temp, _ := io.ReadAll(resp.Body)
		response := responseBody{}
		err := json.Unmarshal(temp, &response)
		_ = err
		if response.Code == "100000" {
			println("抢到了")
			FLAG = 1
		}
	} else {
		println("没抢到")
		return
	}
	println("抢票结束")
}

func waitGroup(request *http.Request) {
	count := 10
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(request *http.Request) {
			sendOrderOrder(request)
			wg.Done() // 也可使用 wg.Add(-1)
		}(request)
	}
	wg.Wait()
}

func generateOptions(start, end int) []string {
	options := make([]string, end-start+1)
	for i := start; i <= end; i++ {
		options[i-start] = fmt.Sprintf("%02d", i)
	}
	return options
}
func main() {

	myApp := app.New()
	myWindow := myApp.NewWindow("ticket")
	HLable := widget.NewLabel("BeginH:")
	startHourW := widget.NewSelectEntry(generateOptions(0, 23))
	startHourW.Text = "16"
	ELable := widget.NewLabel("EndH")
	endHourW := widget.NewSelectEntry(generateOptions(0, 23))
	endHourW.Text = "18"

	YLable := widget.NewLabel("Y:")
	startYW := widget.NewSelectEntry(generateOptions(2024, 2030))
	startYW.Text = "2024"
	MLable := widget.NewLabel("M:")
	startMW := widget.NewSelectEntry(generateOptions(1, 12))
	DLable := widget.NewLabel("D:")
	startDW := widget.NewSelectEntry(generateOptions(1, 31))

	Tag := binding.NewString()
	Tag.Set("waiting")

	content := widget.NewButton("submit", func() {
		sH, _ := strconv.Atoi(startHourW.Text)
		eH, _ := strconv.Atoi(endHourW.Text)
		if sH >= eH || sH < 6 || eH > 23 {
			log.Printf("输入失败重新输入")
		} else {
			Tag.Set("doing")
			beginH = sH
			endH = eH
			date = startYW.Text + "-" + startMW.Text + "-" + startDW.Text
			c := cron.New()
			_ = c.AddFunc("0 */10 * * * *", func() {
				count := 0
				println("________执行_________")
				for FLAG < 1 && count < 2 {
					index := getIndex()
					strId := getOoder(index, beginH, endH)
					request := prepareOrder(index, strId)
					waitGroup(request)
					count++
					println(FLAG)
					if FLAG == 1 {
						Tag.Set("success")
						c.Stop()
					}
				}
			})
			if FLAG != 1 {
				c.Start()
				t := time.NewTimer(time.Minute * 20)
				for {
					select {
					case <-t.C:
						t.Reset(time.Minute * 20)
					}
				}
			}
		}
	})

	myWindow.SetContent(container.New(layout.NewGridLayout(6),
		YLable, startYW, MLable, startMW, DLable, startDW,
		HLable, startHourW, ELable, endHourW, widget.NewLabelWithData(Tag), content,
	))
	myWindow.ShowAndRun()
}
