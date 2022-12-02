package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mahjongSoul/tools"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/nbari/violetear"
)

var CONFIGPATH = "./tools/config.json"
var CONFIG = new(tools.Config)

func initConfig() error {
	file, err := os.Open(CONFIGPATH)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&CONFIG)
	return err
}

func getMahjongURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/mahjong" {
		tmpl, err := template.ParseFiles("./index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			log.Println("create template failed, err:", err)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			return
		}
	} else {
		http.NotFound(w, r)
	}
}

func analyse(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/mahjong/analyse" {
		tmpl, err := template.ParseFiles("./analyse.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			log.Println("create template failed, err:", err)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			return
		}

		err = r.ParseForm()
		if err != nil {
			return
		}
		paras := map[string]string{"url": "", "nickname": "", "jikaze": "", "engine": "mortal"} // 默认mortal引擎
		kaze := map[string]string{"东": "0", "南": "1", "西": "2", "北": "3"}
		for k, v := range r.Form {
			switch k {
			case "url":
				paras["url"] = v[0]
			case "nickname":
				paras["nickname"] = v[0]
			case "jikaze":
				if len(v) != 0 && (v[0] == "东" || v[0] == "南" || v[0] == "西" || v[0] == "北") {
					paras["jikaze"] = kaze[v[0]]
				} else {
					// 返回参数格式错误
					w.WriteHeader(http.StatusBadRequest)
					err = fmt.Errorf("parameter error")
					fmt.Fprintln(w, err)
					log.Println("/analyse: ", err)
					return
				}
			case "engine":
				if len(v) != 0 && (v[0] == "mortal" || v[0] == "akochan") {
					paras["engine"] = v[0]
				} else {
					paras["engine"] = "mortal"
				}
			}
		}
		// 若没有获取到想要的nickname和jikaze参数
		if paras["nickname"] == "" && paras["jikaze"] == "" {
			// 返回参数格式错误
			w.WriteHeader(http.StatusBadRequest)
			err = fmt.Errorf("parameter error")
			fmt.Fprintln(w, err)
			log.Println("/analyse: ", err)
			return
		}
		// 以下检查获取到的url。
		// url格式: "https://game.maj-soul.net/1/?paipu=220926-880ecd12-0b0b-467a-89db-172fe7191263_a57320168"
		reg := regexp.MustCompile(`https://game.maj-soul.net/1/\?paipu=\d{6}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
		if reg == nil {
			log.Println("regexp err")
			return
		}
		res := reg.FindString(paras["url"])
		if res == "" {
			_, err = fmt.Fprintf(w, "不是雀魂牌谱的URL格式！\n")
			if err != nil {
				return
			}
			return
		} else {
			s := strings.Split(paras["url"], "=")
			uuid := strings.Split(s[1], "_")[0]
			_, err = fmt.Fprintf(w, "后台分析中！\n")
			if err != nil {
				return
			}
			paras["jikaze"], err = tools.Comm(CONFIG, uuid, paras["nickname"], paras["jikaze"], paras["engine"], w, r)
			if err != nil {
				fmt.Fprintln(w, err)
				_, err = fmt.Fprintf(w, "分析失败！\n")
				if err != nil {
					return
				}
				return
			} else {
				_, err = fmt.Fprintf(w, "分析成功！3秒后自动跳转...\n")
				if err != nil {
					return
				}
				/*
					// 以POST方式跳转到结果界面
					redirect := `
						<form id="result" action="/mahjong/result" method="post">
							<input type="hidden" name="uuid" value=%s>
							<input type="hidden" name="jikaze" value=%s>
							<input type="hidden" name="engine" value=%s>
						</form>
						<script>setTimeout((() => document.getElementById('result').submit()), 3000);</script>
						<script>alert("分析成功！3秒后自动跳转...");</script>
					`
				*/
				// 以GET方式跳转到结果界面（不可回退至本页，也就是/analyse）
				redirect := `
					<script>setTimeout((() => window.location.replace("/mahjong/result?uuid=%s&jikaze=%s&engine=%s")), 3000);</script>
					<script>alert("分析成功！3秒后自动跳转...");</script>
				`
				fmt.Fprint(w, fmt.Sprintf(redirect, uuid, paras["jikaze"], paras["engine"]))
			}
		}
	} else {
		http.NotFound(w, r)
	}
}

func result(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/mahjong/result" {
		err := r.ParseForm()
		if err != nil {
			return
		}
		paras := map[string]string{"uuid": "", "jikaze": "", "engine": ""}
		for k, v := range r.Form {
			switch k {
			case "uuid":
				if len(v) != 0 && v[0] != "" {
					paras["uuid"] = v[0]
				}
			case "jikaze":
				if len(v) != 0 && v[0] == "0" || v[0] == "1" || v[0] == "2" || v[0] == "3" {
					paras["jikaze"] = v[0]
				} else {
					// 返回参数格式错误
					w.WriteHeader(http.StatusBadRequest)
					err = fmt.Errorf("parameter error")
					fmt.Fprintln(w, err)
					log.Println("/result: ", err)
					return
				}
			case "engine":
				if len(v) != 0 && (v[0] == "mortal" || v[0] == "akochan") {
					paras["engine"] = v[0]
				} else {
					paras["engine"] = "mortal"
				}
			}
		}
		// 若没有获取到想要的uuid或jikaze或engine参数
		if paras["uuid"] == "" || paras["jikaze"] == "" || paras["engine"] == "" {
			// 返回参数格式错误
			w.WriteHeader(http.StatusBadRequest)
			err = fmt.Errorf("parameter error")
			fmt.Fprintln(w, err)
			log.Println("/result: ", err)
			return
		}

		tmpl, err := template.ParseFiles(fmt.Sprintf(CONFIG.ReviewerPath+"/outputs/%s_%s_%s.html", paras["uuid"], paras["jikaze"], paras["engine"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, err)
			log.Println("create template failed, err:", err)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			return
		}
	} else {
		http.NotFound(w, r)
	}
}

func main() {
	err := initConfig()
	if err != nil {
		log.Println(err)
		panic(err)
		return
	}
	router := violetear.New()
	router.HandleFunc("/mahjong", getMahjongURL, "GET")
	router.HandleFunc("/mahjong/analyse", analyse, "POST")
	router.HandleFunc("/mahjong/result", result, "GET,POST")
	err = http.ListenAndServe(":9090", router)
	if err != nil {
		log.Println(err)
		return
	}
}
