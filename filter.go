package wordfilter

import (
	"regexp"
)

// WordType ...
type WordType int

const (
	// WordTypeBlack 敏感词
	WordTypeBlack WordType = iota
	// WordTypeWhitePre 白名单前缀
	WordTypeWhitePre
	// WordTypeWhiteSuf 白名单后缀
	WordTypeWhiteSuf
)

// Filter 敏感词过滤器
type Filter struct {
	black       *Trie //敏感词
	whitePrefix *Trie //白名单(前缀)
	whiteSuffix *Trie //白名单(后缀)
	checkWhite  bool  //是否启用白名单检查
	noise       *regexp.Regexp
}

// New 返回一个敏感词过滤器
func New() *Filter {
	return &Filter{
		black:       NewTrie(),
		whitePrefix: NewTrie(),
		whiteSuffix: NewTrie(),
		checkWhite:  false,
		noise:       regexp.MustCompile(`[\|\s&%$@*]+`),
	}
}

// UpdateNoisePattern 更新去噪模式
func (filter *Filter) UpdateNoisePattern(pattern string) {
	filter.noise = regexp.MustCompile(pattern)
}

// AddWord 添加敏感词
func (filter *Filter) AddWord(typ WordType, words ...string) {
	switch typ {
	case WordTypeBlack:
		filter.black.Add(words...)
	case WordTypeWhitePre:
		for i := range words {
			filter.whitePrefix.Add(filter.overturnString(words[i]))
		}
	case WordTypeWhiteSuf:
		filter.whiteSuffix.Add(words...)
	}
}

// Replace 敏感词替换
func (filter *Filter) Replace(text string, repl rune) string {
	var (
		parent  = filter.black.Root
		current *Node
		runes   = []rune(text)
		left    = 0
		found   bool
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]

		if !found {
			parent = filter.black.Root
			position = left
			left++
			continue
		}

		if current.IsEnd() && left <= position {
			if filter.IsInWhiteList(runes, left, position) {
				left = position + 1
				continue
			}
			for i := left; i <= position; i++ {
				runes[i] = repl
			}
			left = position + 1
		}

		parent = current
	}

	return string(runes)
}

// Filter 过滤敏感词
func (filter *Filter) Filter(text string) string {
	var (
		parent      = filter.black.Root
		current     *Node
		left        = 0
		found       bool
		runes       = []rune(text)
		length      = len(runes)
		resultRunes = make([]rune, 0, length)
	)

	for position := 0; position < length; position++ {
		current, found = parent.Children[runes[position]]

		if !found {
			resultRunes = append(resultRunes, runes[left])
			parent = filter.black.Root
			position = left
			left++
			continue
		}

		if current.IsEnd() {
			if filter.IsInWhiteList(runes, left, position) {
				continue
			}

			left = position + 1
		}
		parent = current
	}

	resultRunes = append(resultRunes, runes[left:]...)
	return string(resultRunes)
}

// Validate 验证字符串是否合法，如不合法则返回false和检测到
// 的第一个敏感词
func (filter *Filter) Validate(text string) (bool, string) {
	const (
		Empty = ""
	)
	var (
		parent  = filter.black.Root
		current *Node
		runes   = []rune(filter.RemoveNoise(text))
		left    = 0
		found   bool
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]

		if !found {
			parent = filter.black.Root
			position = left
			left++
			continue
		}

		if current.IsEnd() && left <= position {
			if filter.IsInWhiteList(runes, left, position) {
				continue
			}

			return false, string(runes[left : position+1])
		}

		parent = current
	}

	return true, Empty
}

// FindIn 判断text中是否含有词库中的词
func (filter *Filter) FindIn(text string) (bool, string) {
	validated, first := filter.Validate(filter.RemoveNoise(text))
	return !validated, first
}

// FindAll 找有所有包含在词库中的词
func (filter *Filter) FindAll(text string) []string {
	var matches []string
	var (
		parent  = filter.black.Root
		current *Node
		runes   = []rune(text)
		length  = len(runes)
		left    = 0
		found   bool
	)

	for position := 0; position < length; position++ {
		current, found = parent.Children[runes[position]]

		if !found {
			parent = filter.black.Root
			position = left
			left++
			continue
		}

		if current.IsEnd() && left <= position {
			if !filter.IsInWhiteList(runes, left, position) {
				matches = append(matches, string(runes[left:position+1]))
			}

			if position == length-1 {
				parent = filter.black.Root
				position = left
				left++
				continue
			}
		}

		parent = current
	}

	if count := len(matches); count > 0 {
		set := make(map[string]struct{})
		for i := range matches {
			_, ok := set[matches[i]]
			if !ok {
				set[matches[i]] = struct{}{}
				continue
			}
			count--
			copy(matches[i:], matches[i+1:])
			i--
		}
		return matches[:count]
	}

	return nil
}

// IsInWhiteList 查白名单
func (filter *Filter) IsInWhiteList(runes []rune, left, position int) bool {
	if filter.checkWhite {
		if left > 0 {
			// 从后往前，查看是否包含白名单前缀
			if ok, _ := filter.FindPrefix(runes[0:position+1], position+1-left); ok {
				return true
			}
		}
		if position < len(runes)-1 {
			// 从前往后，查看是否包含白名单后缀
			if ok, _ := filter.FindSuffix(runes[left:], position+1-left); ok {
				return true
			}
		}
	}
	return false
}

// FindPrefix 查找前缀
func (filter *Filter) FindPrefix(runes []rune, offset int) (bool, string) {
	const (
		Empty = ""
	)
	var (
		parent  = filter.whitePrefix.Root
		current *Node
		found   bool
		l       = len(runes)
		cnt     = 0
	)

	for position := 0; position < l; position++ {
		current, found = parent.Children[runes[l-1-position]]
		if !found {
			// 从后往前找，超过偏移量都没有找到前缀
			if position > offset {
				return false, Empty
			}
			cnt++
		}

		if current.IsEnd() {
			return true, string(runes[l-1-position : l-cnt])
		}

		parent = current
	}

	return false, Empty
}

// FindSuffix 查找后缀
func (filter *Filter) FindSuffix(runes []rune, offset int) (bool, string) {
	const (
		Empty = ""
	)
	var (
		parent  = filter.whiteSuffix.Root
		current *Node
		found   bool
		cnt     = 0
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]
		if !found {
			// 从前往后找，超过偏移量都没有找到后缀
			if position > offset {
				return false, Empty
			}
			cnt++
		}

		if current.IsEnd() {
			return true, string(runes[cnt : position+1])
		}

		parent = current
	}

	return false, Empty
}

// RemoveNoise 去除空格等噪音
func (filter *Filter) RemoveNoise(text string) string {
	return filter.noise.ReplaceAllString(text, "")
}

// SetWhiteFlag 设置白名单启用状态
func (filter *Filter) SetWhiteFlag(isUsing bool) {
	filter.checkWhite = isUsing
}

// 字符串翻转
func (filter *Filter) overturnString(text string) string {
	key := []rune(text)
	l := len(key)
	for j := 0; j < l/2; j++ {
		// 倒序存前缀
		key[j], key[l-1-j] = key[l-1-j], key[j]
	}
	return string(key)
}
