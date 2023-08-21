package mygeecache

type ByteView struct {
	b []byte
}

// 关于如何判断使用值接收者还是引用接收者
//
// 对于ByteView来说，它的成员是引用类型，而引用类型的值是header值，本身就是为了复制而设计的，所以永远都不需要共享一个引用类型的值
func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
