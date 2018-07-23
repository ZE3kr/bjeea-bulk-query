# 北京高考录取结果批量查询工具

本程序可以通过北京教育考试院网站批量查询考生成绩，只需要知道准考证号和考生号。


## 查询单个考生录取结果

```bash
bjeea-bulk-query -i 123456789:12345678901234
```

成功返回结果样例

```
姓名: 张三
准考证号: 123456789
考生号: 12345678901234
本科一批: 北京工业大学 (1049)
专业: 软件工程(实验班) (67)
```

失败返回结果样例 (考生号错误、尚未录取、过期查询或者考试 ID 不匹配)

```
查询失败，请检查准考证号和考生号
准考证号: 111111111
考生号: 22222222222222
```

## 批量查询

```bash
bjeea-bulk-query file.csv
```

本程序会多线程的查询成绩，输出结果顺序可能会有所改变

### `file.csv` 样例

请按照准考证号、考生号创建文件。不符合规则的行会自动跳过。

```CSV
123456789,12345678901234
987654321,43210987654321
111111111,22222222222222
```

返回结果样例

```
姓名: 张三
准考证号: 123456789
考生号: 12345678901234
本科一批: 北京工业大学 (1049)
专业: 软件工程(实验班) (67)

------

姓名: 李四
准考证号: 987654321
考生号: 43210987654321
本科一批: 清华大学 (1023)
专业: 理科试验班类(数理) (10)

------

查询失败，请检查准考证号和考生号
准考证号: 111111111
考生号: 22222222222222
```


## 批量查询 (以 CSV 格式返回)

```bash
bjeea-bulk-query file.csv --csv
```

加入 `--csv` 参数后可以返回 CSV 格式， 成功返回结果样例

```
姓名,准考证号,考生号,大学类型,大学名称,大学代码,专业名称,专业代码,查询状态
张三,12345678,12345678901234,本科一批,北京工业大学,1049,软件工程(实验班),67,成功
李四,87654321,43210987654321,本科一批,清华大学,1023,理科试验班类(数理),10,成功
,111111111,22222222222222,,,0,,0,失败
```

失败返回结果样例