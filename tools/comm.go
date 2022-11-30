package tools

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/Pragmatism0220/majsoul"
	"github.com/Pragmatism0220/majsoul/message"
	"github.com/thedevsaddam/gojsonq/v2"
)

// Majsoul 组合库中的 Majsoul 结构
type Majsoul struct {
	*majsoul.Majsoul
	seat  uint32
	tiles []string
}

// NewMajsoul 创建一个 Majsoul 结构
func newMajsoul() (*Majsoul, error) {
	mj, err := majsoul.New()
	if err != nil {
		return new(Majsoul), err
	}
	mSoul := &Majsoul{Majsoul: mj}
	mSoul.Implement = mSoul // 使用多态实现，如果调用时没有提供外部实现则调用内部的实现，如果没有给 Implement 赋值，则只会调用内部实现
	return mSoul, nil
}

// 判断给定的文件或文件夹是否存在
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func engineReview(conf *Config, uuid string, nickname string, jikaze string, engine string, w http.ResponseWriter, r *http.Request) error {
	// 流式传输
	ctx := r.Context()
	ch := make(chan struct{})

	var cmd *exec.Cmd

	// windows下使用akochan官方打包好的akochan-reviewer-v0.7.1的命令
	// cmd = exec.CommandContext(ctx, "./akochan-reviewer.exe", "-a", jikaze, "--no-open", "--show-rating", "-i", conf.ReviewerPath+"/outputs/"+uuid+".json", "-o", conf.ReviewerPath+"/outputs/"+uuid+"_"+jikaze+"_"+engine+".html")

	// 使用升级后的mjai-reviewer的命令（Linux）
	cmd = exec.CommandContext(ctx, "./mjai-reviewer", "-e", engine, "-a", jikaze, "--no-open", "--show-rating", "-i", conf.ReviewerPath+"/outputs/"+uuid+".json", "-o", conf.ReviewerPath+"/outputs/"+uuid+"_"+jikaze+"_"+engine+".html")

	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		log.Println(err)
		return err
	}
	cmd.Stdout = wPipe
	cmd.Stderr = wPipe
	cmd.Dir = conf.ReviewerPath
	if runtime.GOOS == "windows" {
		cmd.Env = []string{"OMP_NUM_THREADS=8"}
	}

	if err = cmd.Start(); err != nil {
		log.Println(err)
		return err
	}

	var writeError error = nil

	go writeOutput(w, rPipe, &writeError)

	go func(ch chan struct{}) {
		cmd.Wait()
		wPipe.Close()
		ch <- struct{}{}
	}(ch)

	select {
	case <-ch:
	case <-ctx.Done():
		err := ctx.Err()
		log.Printf("Client disconnected: %s\n", err)
		return err
	}
	return writeError
}

func writeOutput(w http.ResponseWriter, input io.ReadCloser, flag *error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Important to make it work in browsers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	in := bufio.NewScanner(input)
	for in.Scan() {
		data := in.Text()
		log.Printf("data: %s\n", data)
		fmt.Fprintf(w, "data: %s\n", data)
		if data[0:6] == "error:" {
			*flag = fmt.Errorf(data)
		}
		flusher.Flush()
	}
	input.Close()
}

// 返回自风对应的下标["0", "1", "2", "3"]，以及错误信息
func Comm(conf *Config, uuid string, nickname string, jikaze string, engine string, w http.ResponseWriter, r *http.Request) (string, error) {
	// 创建outputs缓存文件夹
	path := conf.ReviewerPath + "/outputs/"
	exist, cerr := exists(path)
	if cerr != nil {
		log.Printf("get dir error![%v]\n", cerr)
		return "", cerr
	}
	if !exist {
		cerr = os.Mkdir(path, os.ModePerm)
		if cerr != nil {
			log.Printf("mkdir failed![%v]\n", cerr)
			return "", cerr
		}
	}

	// 判断牌谱json文件是否存在缓存。如果不存在则下载
	exist, cerr = exists(conf.ReviewerPath + "/outputs/" + uuid + ".json")
	if cerr != nil {
		log.Printf("get json file error![%v]\n", cerr)
		return "", cerr
	}
	if !exist {
		var username = conf.Username
		var password = conf.Password
		mSoul, err := newMajsoul()
		if err != nil {
			log.Println(err)
			return "", err
		}
		log.Printf("WebSocket网关: %s", mSoul.ServerAddress.GatewayAddress)

		mSoul.UUID = conf.LoginUUID // 可以直接赋空串，但最好不要经常变。可以写个极小概率赋空串？
		resLogin, err := mSoul.Login(username, password)
		if err != nil {
			log.Println(err)
			return "", err
		}
		if resLogin.Error != nil {
			log.Println(resLogin.Error)
			return "", fmt.Errorf(resLogin.Error.String())
		}

		log.Printf("登录成功！")

		defer mSoul.Logout(mSoul.Ctx, &message.ReqLogout{})

		// 获取具体牌谱内容
		reqGameRecord := message.ReqGameRecord{
			GameUuid:            uuid,
			ClientVersionString: strings.TrimRight("web-"+mSoul.Version.Version, ".w"),
		}
		respGameRecord, err := mSoul.FetchGameRecord(mSoul.Ctx, &reqGameRecord)
		if err != nil {
			return "", err
		}

		majsoulLog := Downloadlog(respGameRecord)

		if err = os.WriteFile(path+uuid+".json", majsoulLog, 0644); err != nil {
			return "", err
		}
	}

	// 以昵称为第一优先，获取昵称对应的自风。执行该段语句的前提是之前保证了服务器端存在牌谱的json文件。
	if nickname != "" {
		res := gojsonq.New().File(conf.ReviewerPath + "/outputs/" + uuid + ".json").From("name").Get()
		hasNickname := false
		for i, e := range res.([]interface{}) {
			if e == nickname {
				jikaze = strconv.Itoa(i)
				hasNickname = true
				break
			}
		}
		if !hasNickname { // 说明没找到对应的昵称
			return "", fmt.Errorf("no corresponding nickname found")
		}
	}

	// 检查服务器上是否有分析结果的缓存
	exist, cerr = exists(conf.ReviewerPath + "/outputs/" + uuid + "_" + jikaze + "_" + engine + ".html")
	if cerr != nil {
		log.Printf("get analysis result html file error![%v]\n", cerr)
		return "", cerr
	}
	if !exist {
		cerr = engineReview(conf, uuid, nickname, jikaze, engine, w, r)
	}

	return jikaze, cerr
}
