package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
	//"github.com/fsnotify/fsnotify"
	"sync"
	"github.com/johnlepikhin/tscollector/transaction"
	"github.com/johnlepikhin/tscollector/config"
	"github.com/johnlepikhin/tscollector/storage"
	"github.com/johnlepikhin/tscollector/tshttp"
)

func saveStats(transaction transaction.Transaction, mutex sync.Mutex) {
	fmt.Println("Save transaction")
	storage.Storage.SaveTransaction(transaction)
	transaction.Cleanup();
	mutex.Unlock()
}

func startStatsSaver (transaction transaction.Transaction) {
	ticker := time.NewTicker(config.Config.SavePeriodMs)
	mutex := sync.Mutex{}
	go func() {
		for {
			select {
			case <- ticker.C:
				mutex.Lock()
				go saveStats(transaction, mutex)
			}
		}
	}()
}

func cmdParser() {
	flag.StringVar(&config.ConfigFile, "config", "", "Path to configuration file")
	flag.StringVar(&config.ValuesFile, "values", "", "Path to values file")
	flag.Parse()

	if (config.ConfigFile == "") {
		panic("Command line -config is required");
	}

	if (config.ValuesFile == "") {
		panic("Command line -values is required");
	}
}

func listen(transaction transaction.Transaction) {
	httpAddHandlerReal := func(auth config.Auth, w http.ResponseWriter, r *http.Request) {
		tshttp.HttpAddHandler(auth, transaction, w, r)
	}

	httpAddOneHandlerReal := func(auth config.Auth, w http.ResponseWriter, r *http.Request) {
		tshttp.HttpAddOneHandler(auth, transaction, w, r)
	}

	httpGetHandlerReal := func(auth config.Auth, w http.ResponseWriter, r *http.Request) {
		tshttp.HttpGetHandler(auth, w, r)
	}

	var wg sync.WaitGroup
	for _, l := range config.Config.Listen {
		wg.Add(1)
		go func(listen config.Listen) {
			defer wg.Done()
			server := http.NewServeMux()
			server.HandleFunc("/add", tshttp.MakeAuthorizedHttpHandler(listen.Auth, httpAddHandlerReal))
			server.HandleFunc("/addone", tshttp.MakeAuthorizedHttpHandler(listen.Auth, httpAddOneHandlerReal))
			server.HandleFunc("/get", tshttp.MakeAuthorizedHttpHandler(listen.Auth, httpGetHandlerReal))
			http.ListenAndServe(listen.Address, server)
		}(l)
	}
	wg.Wait()
}

func main() {
	cmdParser()
	err := config.ConfigParser()
	if err != nil {
		panic(fmt.Sprintf("Cannot read configuration file: %v", err.Error()))
	}
	err = config.ValuesParser()
	if err != nil {
		panic(fmt.Sprintf("Cannot read values file: %v", err.Error()))
	}

	err = storage.StorageParser(config.Config.Storage)
	if err != nil {
		panic(fmt.Sprintf("Cannot read values file: %v", err.Error()))
	}

	var currentTransaction = transaction.NewTransaction()

	startStatsSaver(currentTransaction)
	listen(currentTransaction)
}