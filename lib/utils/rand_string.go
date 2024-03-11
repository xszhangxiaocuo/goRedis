package utils

import (
	"math/rand"
	"time"
)

// r是一个全局的随机数生成器，初始化时使用当前时间作为种子
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// letters是一个包含大小写字母和数字的字符切片，用于生成随机字符串
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// RandString 生成一个长度不超过n的随机字符串
func RandString(n int) string {
	b := make([]rune, n) // 用于存储随机字符
	for i := range b {
		b[i] = letters[r.Intn(len(letters))] // 从letters中随机选取一个字符，并赋值给b的相应位置
	}
	return string(b) // 将rune切片转换为字符串并返回
}

// hexLetters是一个包含十六进制数字的字符切片，用于生成随机十六进制字符串
var hexLetters = []rune("0123456789abcdef")

// RandHexString 生成一个长度不超过n的随机十六进制字符串
func RandHexString(n int) string {
	b := make([]rune, n) // 创建一个长度为n的rune切片，用于存储随机字符
	for i := range b {
		b[i] = hexLetters[r.Intn(len(hexLetters))] // 从hexLetters中随机选取一个字符，并赋值给b的相应位置
	}
	return string(b) // 将rune切片转换为字符串并返回
}

// RandIndex 返回一个长度为size的随机索引切片，用于从切片中随机挑选元素
func RandIndex(size int) []int {
	result := make([]int, size) // 创建一个长度为size的整数切片，用于存储随机索引
	for i := range result {
		result[i] = i // 初始化索引切片，使其包含从0到size-1的连续整数
	}
	rand.Shuffle(size, func(i, j int) { // 使用rand.Shuffle函数打乱索引切片的顺序
		result[i], result[j] = result[j], result[i] // 交换索引切片中的元素
	})
	return result // 返回随机索引切片
}
