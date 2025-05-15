package cachettl

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mapa map[string]Item
	mu   sync.Mutex // защита основной карты при создании, удалении ключей
	key  *KeyLocker //блокировка доступа к конкретным ключам в основной карте
}

type Item struct {
	data       any
	expiration time.Time
}

type KeyLocker struct {
	locks map[string]*sync.Mutex //карта ключей. Ключ - ключ из основной карты, значение - locked или unlocked из мютекса
	mu    sync.Mutex             //защита карты ключей
}

func New() *Cache { //инициализация новой основной карты
	return &Cache{
		mapa: make(map[string]Item),
		key:  NewKeyLocker(),
	}
}

func NewKeyLocker() *KeyLocker { //инициализация новой карты ключей
	return &KeyLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

func (kl *KeyLocker) Lock(key string) { //Блокировка ключа В КАРТЕ КЛЮЧЕЙ, с которым работаем
	//блокируем карту ключей,
	//проверяет наличие в ней искомого и если нет, то создает его.
	//
	kl.mu.Lock()              //блокируем карту ключей
	lock, ok := kl.locks[key] //проверяем, есть ли в ней ключ key
	if !ok {                  //если нет
		lock = &sync.Mutex{} //создаем новый мютекс
		kl.locks[key] = lock //добавляем ключ в карту ключей
	}
	kl.mu.Unlock() //освобождаем карту ключей
	lock.Lock()    //блокируем мютекс ключа, с которым работаем
}

func (kl *KeyLocker) Unlock(key string) { // функция освобождает конкретный ключ в КАРТЕ КЛЮЧЕЙ
	kl.mu.Lock()              //блокируем карту ключей
	lock, ok := kl.locks[key] //достаем из карты мютекс ключа key

	kl.mu.Unlock() //освобождаем карту ключей

	if ok { //если ключсуществует, то
		lock.Unlock() // освобождаем мютекс ключа key
	}
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.key.Lock(key) //блокировка ключа в карте ключей
	defer c.key.Unlock(key)

	c.mu.Lock() //блокировка основной карты
	defer c.mu.Unlock()

	c.mapa[key] = Item{ //добавление нового элемента в основную карту
		data:       value,
		expiration: time.Now().Add(ttl), // абсолютное время протухания значения(время записи + время жизни)
	}
}

func (c *Cache) Get(key string) any { //получение значения по ключу
	c.key.Lock(key) //блокировка ключа в карте ключей
	defer c.key.Unlock(key)

	c.mu.Lock() //блокировка основной карты
	defer c.mu.Unlock()

	if _, ok := c.mapa[key]; !ok { //если ключа нет
		return "записей не найдено"
	}
	if c.mapa[key].expiration.Before(time.Now()) { //если ttl истек
		return "обьект протух"
	}
	return c.mapa[key].data
}

func (c *Cache) Delete(key string) { //удаление элемента из основной мапы
	c.key.Lock(key) //блокировка ключа в карте ключей
	defer c.key.Unlock(key)

	c.mu.Lock() //блокировка основной карты
	defer c.mu.Unlock()

	if _, ok := c.mapa[key]; !ok { //проверка существования ключа key
		fmt.Printf("No record found with key %s\n", key)
		return
	}
	delete(c.mapa, key) //удаление ключа из основной мапы
	fmt.Printf("key %s was deleted\n", key)

}

func (c *Cache) DelExpired() { //удаление устаревших элементов из мапы
	keysToDel := []string{} //слайс для протухших ключей

	c.mu.Lock() //блокировка основной карты

	for key, item := range c.mapa {
		if item.expiration.Before(time.Now()) { //если момент протухания уже прошел
			keysToDel = append(keysToDel, key)

		}
	}
	c.mu.Unlock()

	for _, key := range keysToDel {
		c.key.Lock(key)
		c.mu.Lock()

		delete(c.mapa, key)
		fmt.Printf("элемент %s протух\n", key)

		c.mu.Unlock()
		c.key.Unlock(key)

	}

}
