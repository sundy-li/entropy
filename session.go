package entropy

type ISessionStore interface {
	Get(key string) interface{}
	Set(key string, value interface{})
	Delete(key string)
	Purge()
}

type Session struct {
	SessionId string
	store     ISessionStore
}

func (self *Session) SetSession(key string, value interface{}) {
	self.store.Set(key, value)
}

func (self *Session) GetSession(key string) interface{} {
	return self.store.Get(key)
}

func (self *Session) DeleteSession(key string) {
	self.store.Delete(key)
}

func (self *Session) Purge() {
	self.store.Purge()
}
