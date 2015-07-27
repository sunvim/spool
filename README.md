#Attention
 just for use simple, so  create & rename this
 
 if U has sth. to advise, just tell me
 
# spool
// create a factory() to be used with channel based pool
factory    := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:4000") }

// create a new channel based pool with an initial capacity of 5 and maximum
// capacity of 30. The factory will create 5 initial connections and put it
// into the pool.
p, err := pool.NewChannelPool(5, 30, factory)

// now you can get a connection from the pool, if there is no connection
// available it will create a new one via the factory function.
conn, err := p.Get()

// do something with conn and put it back to the pool by closing the connection
// (this doesn't close the underlying connection instead it's putting it back
// to the pool).
conn.Close()

// close pool any time you want, this closes all the connections inside a pool
p.Close()

// currently available connections in the pool
current := p.Len()
