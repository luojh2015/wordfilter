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
