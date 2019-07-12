package wordfilter

import (
	"regexp"
	"strings"

	"github.com/luojh2015/sego"
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
	// WordTypeSego sego分词词语
	WordTypeSego
)

// Filter 敏感词过滤器
type Filter struct {
	segmenter   *sego.Segmenter
	black       *Trie //敏感词
	whitePrefix *Trie //白名单(前缀)
	whiteSuffix *Trie //白名单(后缀)
	checkWhite  bool  //是否启用白名单检查
	noise       *regexp.Regexp
}

// New 返回一个敏感词过滤器
func New() *Filter {
	return &Filter{
		segmenter:   &sego.Segmenter{},
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

// LoadSegoDic 加载分词词库
func (filter *Filter) LoadSegoDic(path string) {
	filter.segmenter.LoadDictionary(path)
}

// AddWord 添加敏感词
func (filter *Filter) AddWord(typ WordType, word string, frequency int) {
	filter.segmenter.AddWord(word, "", frequency)
	switch typ {
	case WordTypeBlack:
		filter.black.Add(word)
	case WordTypeWhitePre:
		filter.whitePrefix.Add(filter.overturnString(word))
	case WordTypeWhiteSuf:
		filter.whiteSuffix.Add(word)
	}
}

// Replace 敏感词替换
func (filter *Filter) Replace(text string, repl rune) string {
	var (
		runes    = []rune(text)
		position = 0
	)

	segs := filter.segmenter.Segment([]byte(text))
	for i := range segs {
		word := []rune(segs[i].Token().Text())
		position += len(word)
		// 当前分词在黑名单中 且 不含白名单前缀或后缀
		if filter.IsIn(WordTypeBlack, string(word)) && !filter.IsInWhiteList(text, position-len(word), position-1) {
			for j := position - len(word); j < position; j++ {
				runes[j] = repl
			}
		}
	}

	return string(runes)
}

// Filter 过滤敏感词
func (filter *Filter) Filter(text string) string {
	var (
		runes       = []rune(text)
		resultRunes = make([]rune, 0, len(runes))
		position    = 0
	)

	segs := filter.segmenter.Segment([]byte(text))
	for i := range segs {
		word := []rune(segs[i].Token().Text())
		position += len(word)
		// 当前分词在黑名单中 且 不含白名单前缀或后缀
		if filter.IsIn(WordTypeBlack, string(word)) && !filter.IsInWhiteList(text, position-len(word), position-1) {
			continue
		}
		resultRunes = append(resultRunes, word...)
	}

	return string(resultRunes)
}

// Validate 验证字符串是否合法，如不合法则返回false和检测到
// 的第一个敏感词
func (filter *Filter) Validate(text string) (bool, string) {
	const (
		Empty = ""
	)
	var (
		position = 0
	)

	segs := filter.segmenter.Segment([]byte(text))
	for i := range segs {
		word := []rune(segs[i].Token().Text())
		position += len(word)
		// 当前分词在黑名单中 且 不含白名单前缀或后缀
		if filter.IsIn(WordTypeBlack, string(word)) && !filter.IsInWhiteList(text, position-len(word), position-1) {
			return false, string(word)
		}
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
	var (
		matches  []string
		position = 0
	)

	segs := filter.segmenter.Segment([]byte(text))
	for i := range segs {
		word := []rune(segs[i].Token().Text())
		position += len(word)
		// 当前分词在黑名单中 且 不含白名单前缀或后缀
		if filter.IsIn(WordTypeBlack, string(word)) && !filter.IsInWhiteList(text, position-len(word), position-1) {
			matches = append(matches, string(word))
		}
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
func (filter *Filter) IsInWhiteList(src string, left, position int) bool {
	runes := []rune(strings.ToLower(src))
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
		root    = filter.whitePrefix.Root
		current *Node
		found   bool
		l       = len(runes)
		cnt     = 0
	)

	for position, parent := 0, root; position < l; position++ {
		current, found = parent.Children[runes[l-1-position]]
		if !found {
			// 从后往前找，超过偏移量都没有找到前缀
			if position > offset {
				return false, Empty
			}
			parent = root
			position = cnt
			cnt++
			continue
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
		root    = filter.whiteSuffix.Root
		current *Node
		found   bool
		cnt     = 0
	)

	for position, parent := 0, root; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]
		if !found {
			// 从前往后找，超过偏移量都没有找到后缀
			if position > offset {
				return false, Empty
			}
			parent = root
			position = cnt
			cnt++
			continue
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

// IsIn ...
func (filter *Filter) IsIn(typ WordType, word string) bool {
	var (
		root    *Node
		current *Node
		found   bool
		runes   []rune
	)
	switch typ {
	case WordTypeBlack:
		root = filter.black.Root
		runes = []rune(strings.ToLower(word))
	case WordTypeWhitePre:
		root = filter.whitePrefix.Root
		runes = []rune(strings.ToLower(filter.overturnString(word)))
	case WordTypeWhiteSuf:
		root = filter.whiteSuffix.Root
		runes = []rune(strings.ToLower(word))
	default:
		return false
	}
	for position, parent := 0, root; position < len(runes); position++ {
		current, found = parent.Children[runes[position]]
		if !found {
			break
		}

		// 搜索路径结束在字符串末尾时，该词语在此列表中
		if current.IsEnd() && position == len(runes)-1 {
			return true
		}

		parent = current
	}
	return false
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
