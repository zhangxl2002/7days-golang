package mygeecache

import (
	"fmt"
	"log"
	"sync"
)

// Group代表着一种资源,封装了从缓存获取和直接从来源getter获取两种途径
type Group struct {
	name string
	// 直接获取资源的方式交由用户来决定
	getter    Getter
	mainCache cache
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	// RWMutex支持并发读取,用于保护groups
	mu sync.RWMutex
	// groups是所有资源的集合
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		// 没有提供直接获取值的方法，无法继续
		panic("nil Getter")
	}
	// 因为涉及groups的写，所以使用的是Lock()而不是RLock()
	mu.Lock()
	defer mu.Unlock()
	// 参考go逃逸分析，g逃逸到了NewGroup之外，所以是新创建的Group是分配在堆上的
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	// 由于只涉及groups的读，所以只要加读锁即可
	mu.RLock()
	defer mu.RLock()
	g := groups[name]
	return g
}

// Group暴露给外界的获取值的接口
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 从缓存中顺利获得
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[myGeeCache] hit")
		return v, nil
	}

	// 尝试通过其他方式获得
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// g.getter.Get返回值的类型是[]byte
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	log.Println("debug")
	// 直接获得后，需要更新缓存
	g.populateCache(key, value)
	log.Println("debug")
	return value, nil
}

// 更新缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
