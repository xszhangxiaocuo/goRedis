package utils

// ToCmdLine 将字符串切片转换为字节切片的切片
func ToCmdLine(cmd ...string) [][]byte {
	args := make([][]byte, len(cmd)) // 创建一个长度等于cmd长度的二维字节切片
	for i, s := range cmd {
		args[i] = []byte(s) // 将每个字符串转换为字节切片
	}
	return args // 返回转换后的二维字节切片
}

// ToCmdLine2 将命令名和字符串类型的参数转换为二维字节切片
func ToCmdLine2(commandName string, args ...string) [][]byte {
	result := make([][]byte, len(args)+1) // 创建一个长度为参数个数加1的二维字节切片
	result[0] = []byte(commandName)       // 将命令名转换为字节切片并放在第一个位置
	for i, s := range args {
		result[i+1] = []byte(s) // 将每个参数转换为字节切片
	}
	return result // 返回转换后的二维字节切片
}

// ToCmdLine3 将命令名和字节切片类型的参数转换为二维字节切片
func ToCmdLine3(commandName string, args ...[]byte) [][]byte {
	result := make([][]byte, len(args)+1) // 创建一个长度为参数个数加1的二维字节切片
	result[0] = []byte(commandName)       // 将命令名转换为字节切片并放在第一个位置
	for i, s := range args {
		result[i+1] = s // 直接将字节切片参数放入相应位置
	}
	return result // 返回转换后的二维字节切片
}

// Equals 检查两个值是否相等
func Equals(a interface{}, b interface{}) bool {
	sliceA, okA := a.([]byte) // 尝试将a转换为字节切片
	sliceB, okB := b.([]byte) // 尝试将b转换为字节切片
	if okA && okB {
		return BytesEquals(sliceA, sliceB) // 如果都是字节切片，则使用BytesEquals比较
	}
	return a == b // 否则直接比较两个值
}

// BytesEquals 检查两个字节切片是否相等
func BytesEquals(a []byte, b []byte) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false // 一个为nil另一个不为nil，则不相等
	}
	if len(a) != len(b) {
		return false // 长度不同，则不相等
	}
	size := len(a)
	for i := 0; i < size; i++ {
		if a[i] != b[i] {
			return false // 有任何一个元素不相等，则不相等
		}
	}
	return true // 所有元素都相等，则相等
}

// ConvertRange 将Redis的索引范围转换为Go切片的索引范围
// -1 => size-1
// 区间为闭区间[0, 10]转换为左闭右开区间[0, 9)
// 超出边界的值转换为最大内部边界 [size, size+1] => [-1, -1]
func ConvertRange(start int64, end int64, size int64) (int, int) {
	if start < -size {
		return -1, -1 // 开始索引小于-size，则返回无效区间
	} else if start < 0 {
		start = size + start // 将负数索引转换为正数索引
	} else if start >= size {
		return -1, -1 // 开始索引大于等于size，则返回无效区间
	}
	if end < -size {
		return -1, -1 // 结束索引小于-size，则返回无效区间
	} else if end < 0 {
		end = size + end + 1 // 将负数索引转换为正数索引，并加1
	} else if end < size {
		end = end + 1 // 结束索引小于size，则加1
	} else {
		end = size // 结束索引大于等于size，则设为size
	}
	if start > end {
		return -1, -1 // 开始索引大于结束索引，则返回无效区间
	}
	return int(start), int(end) // 返回转换后的区间
}
