package wildcard

const (
	normal     = iota // 普通字符
	all               // * 	匹配任意长度的字符串
	anys              // ? 	匹配任意单个字符
	setSymbol         // [] 	匹配集合中的任意字符
	rangSymbol        // [a-b]	匹配范围内的任意字符
	negSymbol         // [^a]	匹配不在集合中的任意字符
)

type item struct {
	character byte
	set       map[byte]bool
	typeCode  int
}

func (i *item) contains(c byte) bool {
	if i.typeCode == setSymbol {
		_, ok := i.set[c]
		return ok
	} else if i.typeCode == rangSymbol {
		if _, ok := i.set[c]; ok {
			return true
		}
		var (
			_min uint8 = 255
			_max uint8 = 0
		)
		for k := range i.set {
			if _min > k {
				_min = k
			}
			if _max < k {
				_max = k
			}
		}
		return c >= _min && c <= _max
	} else {
		_, ok := i.set[c]
		return !ok
	}
}

// Pattern 表示一个通配符模式
type Pattern struct {
	items []*item
}

// CompilePattern 将通配符字符串转换为 Pattern 对象
func CompilePattern(src string) *Pattern {
	items := make([]*item, 0)
	escape := false
	inSet := false
	var set map[byte]bool
	for _, v := range src {
		c := byte(v)
		if escape {
			items = append(items, &item{typeCode: normal, character: c})
			escape = false
		} else if c == '*' {
			items = append(items, &item{typeCode: all})
		} else if c == '?' {
			items = append(items, &item{typeCode: anys})
		} else if c == '\\' {
			escape = true
		} else if c == '[' {
			if !inSet {
				inSet = true
				set = make(map[byte]bool)
			} else {
				set[c] = true
			}
		} else if c == ']' {
			if inSet {
				inSet = false
				typeCode := setSymbol
				if _, ok := set['-']; ok {
					typeCode = rangSymbol
					delete(set, '-')
				}
				if _, ok := set['^']; ok {
					typeCode = negSymbol
					delete(set, '^')
				}
				items = append(items, &item{typeCode: typeCode, set: set})
			} else {
				items = append(items, &item{typeCode: normal, character: c})
			}
		} else {
			if inSet {
				set[c] = true
			} else {
				items = append(items, &item{typeCode: normal, character: c})
			}
		}
	}
	return &Pattern{
		items: items,
	}
}

// IsMatch 检查一个字符串是否匹配 Pattern 对象
func (p *Pattern) IsMatch(s string) bool {
	if len(p.items) == 0 {
		return len(s) == 0
	}
	m := len(s)
	n := len(p.items)
	table := make([][]bool, m+1) // 使用动态规划的方式创建一个二维数组，其中 table[i][j] 表示 s 的前 i 个字符是否与 Pattern 的前 j 个 item 匹配。
	for i := 0; i < m+1; i++ {
		table[i] = make([]bool, n+1)
	}
	table[0][0] = true
	for j := 1; j < n+1; j++ {
		table[0][j] = table[0][j-1] && p.items[j-1].typeCode == all
	}
	for i := 1; i < m+1; i++ {
		for j := 1; j < n+1; j++ {
			if p.items[j-1].typeCode == all {
				table[i][j] = table[i-1][j] || table[i][j-1]
			} else {
				table[i][j] = table[i-1][j-1] &&
					(p.items[j-1].typeCode == anys ||
						(p.items[j-1].typeCode == normal && uint8(s[i-1]) == p.items[j-1].character) ||
						(p.items[j-1].typeCode >= setSymbol && p.items[j-1].contains(s[i-1])))
			}
		}
	}
	return table[m][n]
}
