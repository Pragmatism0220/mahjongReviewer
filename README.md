# mahjongReviewer
[mjai-reviewer](https://github.com/Equim-chan/mjai-reviewer)的前后端程序。

## 它是什么
这是一个[mjai-reviewer](https://github.com/Equim-chan/mjai-reviewer)的前后端程序。界面非常简单。在浏览器端输入雀魂的牌谱网址，即可在服务器端调用akochan或Mortal引擎对牌谱进行分析。分析过程会通过流式传输进行实时显示，并在分析结束后自动跳转至结果界面。

> 其中，主界面是扒自[麻将 AI 工具 - 检讨牌谱，派遣个室](https://mjai.ekyu.moe/zh-cn.html)，界面UI版权归原作者所有。这里仅为我个人为了尽可能地模仿、复现而进行的学习，仅供学习使用。如有侵犯，可随时联系我删除。

演示过程如下图所示，依次是主界面登录界面、主界面、分析界面、结果界面。用户可以选择从牌谱链接中自动识别玩家、或者选择自定义指定玩家进行分析：
![主界面登录界面](/images/verify.png)
![主界面](/images/mahjong.png)
![分析界面](/images/analyse.png)
![结果界面](/images/result.png)

### 环境依赖与相关项目
* go1.19.1 环境开发
* Windows 10/Linux理论皆可
* 依赖见`go.mod`
  * 其中用到了与雀魂的通信连接。魔改了项目[majsoul](https://github.com/constellation39/majsoul)。
  * `tools/downloadlog.go`是基于[downloadlog.js](https://gist.githubusercontent.com/Equim-chan/875a232a2c1d31181df8b3a8704c3112/raw/a0533ae7a0ab0158ca9ad9771663e94b82b61572/downloadlogs.js)用go重写实现的。

## 如何使用
首先需保证自己已正确安装并配置go的环境，并开启go mod支持。

其次请务必已正确配置`mjai-reviewer`。详情见[官方项目](https://github.com/Equim-chan/mjai-reviewer)。

利用git克隆该项目到本地：
```
git clone https://github.com/Pragmatism0220/mahjongReviewer.git
```
克隆之后，请配置`tools/config.json`（配置文件）。其格式如下：
```json
{
  "username": "example@email.com",  // 雀魂小号邮箱地址
  "password": "",                   // 雀魂小号密码
  "loginUUID": "",                  // 形如"bbd6p84-oe7u-t4qr-tteb-iwar77s63donn"，可以留空，但最好还是填上
  "reviewerPath": ""                // reviewer路径。比如在linux下：/home/user/some/dir/mjai-reviewer/
}
```
**请注意，由于本项目调用了go-ping包。因此，如果你处在Linux环境下，则需要执行以下命令：（[原因在此](https://github.com/go-ping/ping#linux)）**
```shell
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```
之后，只需要**编译运行`mahjong.go`文件**即可。执行过程中，缓存文件夹会生成：`mjai-reviewer/outputs/`。

在浏览器端输入：
```
http://localhost:9090/mahjong
```
即可进行访问。

## 开源许可证
[Mozilla Public License 2.0](https://github.com/Pragmatism0220/mahjongReviewer/blob/main/LICENSE)