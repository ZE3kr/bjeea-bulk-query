package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Examinee struct {
	name           string
	examNo         int
	examineeNo     int64
	university     string
	universityNo   int
	major          string
	majorNo        int
	universityType string
	failed         bool
}

var (
	inputExaminee = kingpin.Flag("examinee", "考生信息, 格式 [准考证号,考生号]").Short('i').String()
	inputId       = kingpin.Flag("examid", "考试 ID (表单中的 examID, 默认为 2018 年北京市高招录取结果)").Default("4865").Short('t').Int()
	filename      = kingpin.Arg("file.csv", "准考证号与考生号的 CSV 文件").String()
	outCSV        = kingpin.Flag("csv", "以 CSV 格式输出").Bool()
)

func main() {
	kingpin.Parse()
	kingpin.CommandLine.HelpFlag.Short('h')

	if *inputExaminee == "" {
		if *filename == "" || !strings.HasSuffix(*filename, ".csv") {
			kingpin.FatalUsage("参数错误, 请提供文件名或考生信息, 文件必须是 .csv 后缀\n")
		}
		if rawBytes, err := ioutil.ReadFile(os.Args[1]); err != nil {
			log.Fatal(err)
		} else {
			examinees := parseExaminees(string(rawBytes))
			if *outCSV {
				fmt.Println("姓名,准考证号,考生号,大学类型,大学名称,大学代码,专业名称,专业代码,查询状态")
				for _, examinee := range getExamineesDetail(examinees) {
					fmt.Println(examinee.List())
				}
			} else {
				for i, examinee := range getExamineesDetail(examinees) {
					fmt.Println(examinee)
					if i != len(examinees)-1 {
						fmt.Print("------\n\n")
					}
				}
			}

		}
	} else if inputExamineeSlice := strings.Split(*inputExaminee, ","); len(inputExamineeSlice) == 2 {
		var examinee Examinee
		examinee.examNo, _ = strconv.Atoi(strings.TrimSpace(inputExamineeSlice[0]))
		examinee.examineeNo, _ = strconv.ParseInt(strings.TrimSpace(inputExamineeSlice[1]), 10, 64)
		fmt.Print(getExamineeDetail(examinee))
	} else {
		kingpin.FatalUsage("参数错误, 请提供文件名或考生信息. 准考证号应为 9 位, 考生号应为 14 位\n")
	}
}

func parseExaminees(data string) (Examinees []Examinee) {
	var (
		Examinee Examinee
		err      error
	)
	for _, line := range strings.Split(data, "\n") {
		lineSlice := strings.Split(line, ",")
		if line == "" || len(lineSlice) < 2 {
			continue
		} else if examNo, err := strconv.Atoi(strings.TrimSpace(lineSlice[0])); err == nil {
			Examinee.examNo = examNo
			if examineeNo, err := strconv.ParseInt(strings.TrimSpace(lineSlice[1]), 10, 64); err == nil {
				Examinee.examineeNo = examineeNo
				Examinees = append(Examinees, Examinee)
			}
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return Examinees
}

func getExamineesDetail(examinees []Examinee) (newExaminees []Examinee) {
	examineesChannel := make(chan Examinee)
	for _, examinee := range examinees {
		go func(examinee Examinee) {
			examineeDetail := getExamineeDetail(examinee)
			if examineeDetail.name == "" {
				examinee.failed = true
			}
			examineesChannel <- examineeDetail
		}(examinee)
	}

	for len(newExaminees) < len(examinees) {
		examinee := <-examineesChannel
		newExaminees = append(newExaminees, examinee)
	}

	return newExaminees
}

func getExamineeDetail(examinee Examinee) Examinee {
	var urlQuery string
	urlQuery = "examNo=" + fmt.Sprintf("%09d", examinee.examNo) + "&examinneNo=" + fmt.Sprintf("%014d", examinee.examineeNo) + "&examId=" + fmt.Sprint(*inputId)
	resp, err := http.Post("http://query.bjeea.cn/queryService/rest/admission/110",
		"application/x-www-form-urlencoded; charset=utf-8",
		strings.NewReader(urlQuery))
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		panic(err)
	}
	if jsonMap["enrollList"] != nil {
		jsonMap, _ = jsonMap["enrollList"].([]interface{})[0].(map[string]interface{})
		examinee.university, _ = jsonMap["GRADE11"].(string)
		examinee.universityNo, _ = strconv.Atoi(jsonMap["GRADE10"].(string))
		examinee.major, _ = jsonMap["GRADE13"].(string)
		examinee.majorNo, _ = strconv.Atoi(jsonMap["GRADE12"].(string))
		examinee.name, _ = jsonMap["NAME"].(string)
		examinee.universityType = jsonMap["GRADE8"].(string)
	} else {
		examinee.failed = true
	}

	return examinee
}

func (examinee Examinee) String() (formatted string) {
	if examinee.failed {
		formatted = "查询失败, 请检查准考证号和考生号\n"
	}
	if examinee.name != "" {
		formatted = "姓名: " + examinee.name + "\n"
	}
	if examinee.examNo > 0 {
		formatted += "准考证号: " + fmt.Sprintf("%09d", examinee.examNo) + "\n"
	}
	if examinee.examineeNo > 0 {
		formatted += "考生号: " + fmt.Sprintf("%014d", examinee.examineeNo) + "\n"
	}
	if examinee.universityNo > 0 && examinee.universityType != "" && examinee.university != "" {
		formatted += examinee.universityType + ": " + examinee.university + " (" + fmt.Sprint(examinee.universityNo) + ")\n"
	}
	if examinee.majorNo > 0 && examinee.major != "" {
		formatted += "专业: " + examinee.major + " (" + fmt.Sprint(examinee.majorNo) + ")\n"
	}
	return formatted
}

func (examinee Examinee) List() (formatted string) {
	formatted = examinee.name
	formatted += "," + fmt.Sprintf("%09d", examinee.examNo)
	formatted += "," + fmt.Sprintf("%014d", examinee.examineeNo)
	formatted += "," + examinee.universityType + "," + examinee.university + "," + fmt.Sprint(examinee.universityNo)
	formatted += "," + examinee.major + "," + fmt.Sprint(examinee.majorNo)
	if examinee.failed {
		formatted += ",失败"
	} else {
		formatted += ",成功"
	}
	return formatted
}
