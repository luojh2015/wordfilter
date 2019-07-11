# wordfilter
敏感词查找,验证,过滤和替换 🤓 FindAll, Validate, Filter and Replace words. 

Usage:

```go
package main

import (
	"fmt"
	"github.com/luojh2015/wordfilter"
)

func main() {
	filter := wordfilter.New()
	filter.AddWord("测试")
	str:="仅仅只是想要测试一下"
	str1:="仅仅只是想要测x试一下"

	fmt.Println(filter.Filter(str))
	fmt.Println(filter.Replace(str, '*'))
	fmt.Println(filter.FindIn(str))
	fmt.Println(filter.Validate(str))
	fmt.Println(filter.FindAll(str))

	fmt.Println(filter.FindIn(str1))
	filter.UpdateNoisePattern(`x`)
	fmt.Println(filter.FindIn(str1))
	fmt.Println(filter.Validate(str1))
}
```
