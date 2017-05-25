package tshttp

import
(
	"net/http"
	"encoding/json"
	"tscollector/transaction"
	"tscollector/config"
	"tscollector/httpjson"
	"tscollector/storage"
	"time"
	"strconv"
)

type Transaction map[config.MeasureKey]string

type TimeSeries map[int64]Transaction

type Response struct {
	Status int
	TimeSeries TimeSeries
}

func HttpAddHandler (auth config.Auth, sharedTransaction transaction.Transaction, w http.ResponseWriter, r *http.Request) {
	if !auth.AllowAddValues {
		httpjson.HttpJsonError(w, 1, "No write access")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var values transaction.InputValues
	err := decoder.Decode(&values)
	if err != nil {
		httpjson.HttpJsonError(w, 1, "Invalid input JSON")
		return
	}
	defer r.Body.Close()
	err = sharedTransaction.AddPlainValues(auth, values)
	if err != nil {
		httpjson.HttpJsonError(w, 1, err.Error())
		return
	}
	httpjson.HttpJsonOK(w)
}

func HttpAddOneHandler (auth config.Auth, sharedTransaction transaction.Transaction, w http.ResponseWriter, r *http.Request) {
	if !auth.AllowAddValues {
		httpjson.HttpJsonError(w, 1, "No write access")
		return
	}

	queryKey, ok := r.URL.Query()["key"]
	if !ok {
		httpjson.HttpJsonError(w, 1, "'key' is required")
		return
	}

	queryValue, ok := r.URL.Query()["value"]
	if !ok {
		httpjson.HttpJsonError(w, 1, "'value' is required")
		return
	}

	var qType = "unknown"
	queryType, ok := r.URL.Query()["type"]
	if ok {
		qType = queryType[0]
	}

	err := sharedTransaction.AddPlainValue(auth, config.MeasureKey(queryKey[0]), transaction.InputValue{queryValue[0], qType})
	if err != nil {
		httpjson.HttpJsonError(w, 1, err.Error())
		return
	}
	httpjson.HttpJsonOK(w)
}

func HttpGetHandler (auth config.Auth, w http.ResponseWriter, r *http.Request) {
	if !auth.AllowReadValues {
		httpjson.HttpJsonError(w, 1, "No read access")
		return
	}

	var start time.Time
	var end time.Time

	queryStart, ok := r.URL.Query()["start"]
	if !ok {
		start = time.Now().Add(-time.Hour)
	} else {
		start_int, err := strconv.ParseInt(queryStart[0], 10, 64)
		if err != nil {
			httpjson.HttpJsonError(w, 1, "Invalid value of URI parameter 'start': " + err.Error())
			return
		}
		start = time.Unix(0,  start_int*time.Millisecond.Nanoseconds())
	}

	queryEnd, ok := r.URL.Query()["end"]
	if !ok {
		end = time.Now()
	} else {
		end_int, err := strconv.ParseInt(queryEnd[0], 10, 64)
		if err != nil {
			httpjson.HttpJsonError(w, 1, "Invalid value of URI parameter 'end': " + err.Error())
			return
		}
		end = time.Unix(0,  end_int*time.Millisecond.Nanoseconds())
	}

	if end.UnixNano() <= start.UnixNano() {
		httpjson.HttpJsonError(w, 1, "URI parameter 'end' must be greater than 'start'")
		return
	}

	if end.Sub(start) > config.Config.MaxGetInterval {
		httpjson.HttpJsonError(w, 1, "Time interval between values of URI parameters 'start' and 'end' is too big. Max interval is " +
			strconv.FormatUint(uint64(config.Config.MaxGetInterval.Seconds()), 10) + " seconds")
		return
	}

	timeSeries, err := storage.Storage.LoadTimeSeries(start, end)
	if err != nil {
		httpjson.HttpJsonError(w, 1, "Cannot load time series: " + err.Error())
		return
	}

	var responseTimeSeries = make(TimeSeries)
	for timeStamp, storageTransaction := range timeSeries {
		var resultTransaction = make(Transaction)
		for key, transactionValue := range storageTransaction.GetValues() {
			resultTransaction[key] = transactionValue.Printable()
		}
		responseTimeSeries[timeStamp.UnixNano() / time.Millisecond.Nanoseconds()] = resultTransaction
	}

	var response = Response{
		Status: 0,
		TimeSeries: responseTimeSeries,
	}

	httpjson.HttpJsonResponse(w, func(jsonw *json.Encoder) {
		jsonw.Encode(response)
	})
}

func MakeAuthorizedHttpHandler(auth []config.Auth, handler func(config.Auth, http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		received_username, received_password, _ := r.BasicAuth();

		unauthorized := func() {
			w.Header().Set("WWW-Authenticate", `Basic realm="tscollector"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		for _, auth := range auth {
			if auth.Username == received_username {
				if auth.Password == received_password {
					handler(auth, w, r)
					return
				}
			} else {
				break
			}
		}

		unauthorized()
	}
}

