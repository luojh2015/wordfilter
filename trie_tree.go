package wordfilter

// Trie 短语组成的Trie树.
type Trie struct {
	Root *Node
}

// Node Trie树上的一个节点.
type Node struct {
	isRoot    bool
	isEnd     bool
	Character rune
	Children  map[rune]*Node
}

// NewTrie 新建一棵Trie
func NewTrie() *Trie {
	return &Trie{
		Root: &Node{
			isRoot:   true,
			Children: make(map[rune]*Node, 0),
		},
	}
}

// Add 添加若干个词
func (tree *Trie) Add(words ...string) {
	for _, word := range words {
		tree.add(word)
	}
}

func (tree *Trie) add(word string) {
	var current = tree.Root
	var runes = []rune(word)
	for position := 0; position < len(runes); position++ {
		r := runes[position]
		if next, ok := current.Children[r]; ok {
			current = next
		} else {
			newNode := NewNode(r)
			current.Children[r] = newNode
			current = newNode
		}
		if position == len(runes)-1 {
			current.isEnd = true
		}
	}
}

// Delete 删除一个词
func (tree *Trie) Delete(word string) {
	type D struct {
		c *Node
		p *Node
	}
	var current = tree.Root
	var runes = []rune(word)
	var path []D
	for position := 0; position < len(runes); position++ {
		r := runes[position]
		if next, ok := current.Children[r]; ok {
			path = append(path, D{c: next, p: current})
			current = next
		} else {
			// 没有匹配的path
			return
		}
	}
	// 从后往前删除节点
	for position := len(path) - 1; position >= 0; position-- {
		if position != len(path)-1 {
			if path[position].c.IsEnd() {
				// path中存在短词语
				return
			}
		} else if path[position].c.IsEnd() {
			// path最后一个节点删除结束标志
			path[position].c.isEnd = false
		}
		if path[position].c.IsLeaf() {
			// 从后往前，删掉叶子节点
			delete(path[position].p.Children, path[position].c.Character)
			continue
		}
	}
}

// NewNode 新建子节点
func NewNode(character rune) *Node {
	return &Node{
		Character: character,
		Children:  make(map[rune]*Node, 0),
	}
}

// IsLeaf 判断是否叶子节点
func (node *Node) IsLeaf() bool {
	return len(node.Children) == 0
}

// IsRoot 判断是否为根
func (node *Node) IsRoot() bool {
	return node.isRoot
}

// IsEnd 判断是否结束
func (node *Node) IsEnd() bool {
	return node.isEnd
}
