package daemon

import "sync"

var wg sync.WaitGroup
var mutex = &sync.Mutex{}
