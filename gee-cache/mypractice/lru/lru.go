package lru

import (
	"container/list"
)

// lru 缓存
// Cache首字母大写,表示可以被其他包访问
type Cache struct {
	maxBytes int64
	nbytes   int64
	//lru中的链表，其中存储着键值对的数据
	ll *list.List

	//lru中的map，从键映射到链表上的位置
	cache map[string]*list.Element

	//驱逐出某个键值对时的回调
	onEvicted func(key string, value Value)
}

// 键值对,链表的Element中的数据类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//更新到链表头部
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// 更新占用空间
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		//更新值
		kv.value = value
	} else {
		// 创建新节点加入到链表中,只需要提供Element的Value就好了
		ele := c.ll.PushFront(&entry{key, value})
		// 更新map
		c.cache[key] = ele
		// 更新占用空间
		c.nbytes += int64(len(key) + value.Len())
	}
	// 超出空间上限淘汰
	// 特判,防止c.maxBytes异常导致程序死循环
	for c.maxBytes > 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	// 链表长度为0时，返回的会是nil（不会返回root）
	if ele != nil {
		// 从链表中删除
		c.ll.Remove(ele)
		// 从map中删除
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		// 更新占有的空间
		c.nbytes -= int64(len(kv.key) + kv.value.Len())
		// 调用注册的回调
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	// 直接return的时候value和ok都是默认值
	return
}
func (c *Cache) Len() int {
	return c.ll.Len()
}
