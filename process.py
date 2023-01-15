s = ""
with open("index.txt") as f:
    line = f.readline()
    while line:
        s += line.split(":", 1)[0] + ", "
        line = f.readline()   
print(s)

"""
本代码实现了逐行读取本地txt文件，并将其输出为JavaScript数组的形式。
可以用来在浏览器端console台爬取雀魂（https://game.maj-soul.com/）人物、役种等函数映射输出。

样例JavaScript代码如下（在console运行）：
var s = "";
var idx = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 1000, 1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010, 1011, 1012, 1013, 1014, 1015, 1016, 1017, 1018, 1019, 1020, 1021];
for (const v of idx) {
    s += ('\"' + v + '\": {\"' + cfg.fan.fan.map_[v].name_jp + '\", \"' + cfg.fan.fan.map_[v].name_jp + '\", \"' + cfg.fan.fan.map_[v].name_en + '\"},\n');
}
console.log(s);
"""
