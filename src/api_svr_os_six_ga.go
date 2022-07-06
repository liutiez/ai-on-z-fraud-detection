package main

import (
	//REST API
	"encoding/json"
	// "errors"
	"io/ioutil"
	"net/http"
	"fmt"
	"strconv"
	//Redis
	"time"
	"github.com/gomodule/redigo/redis"
	//
	"strings"
	//
	"bytes"
	//
	"encoding/csv"
	"os"
	"bufio"
	"io"
	"log"
	"plugin"
)

//Be used to save REST API Request JSON
type Tran struct {
	Tx_index  int `json:"tx_index"`
	Tx_json string `json:"tx_json"` //Optional
}

//Be used to save Predict JSON returned from TFS
type Predict struct {
	Predictions [7][1][1]float64 `json:"predictions"`
}

// Create Test Case Dict
// Read CSV format test case data from file, 
// then transfer CSV format into JSON format and save into test_case_dict.
// Key is Index , Value is JSON string of test case data. 
var test_case_dict = make(map[int]string)
func Create_test_case_dict(csv_row_dict map[int]string) {
	csvFile, _ := os.Open("test_220_100k_os.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	defer csvFile.Close()
	//Remove Header
	line, error := reader.Read()
	//Dict of Test Case data
	//csv_row_dict := make(map[int]string)
	//Read the CSV Body
	for {
		line, error = reader.Read()
		//
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		//
		type Input_Data struct {
			//Empty='missing_value'
			Merchant_State string
			//Empty=0.0
			Zip float64
			Merchant_Name string
			Merchant_City string
			MCC int
			Use_Chip string
			//Empty='missing_value'
			Errors string
			Year int
			Month int
			Day int
			Time string
			Amount string
		}
		//
		var csv_row Input_Data
		csv_row_index,_ := 	strconv.Atoi(line[0])
		csv_row.Year,_ 	= 	strconv.Atoi(line[3])
		csv_row.Month,_ = 	strconv.Atoi(line[4])
		csv_row.Day,_ 	= 	strconv.Atoi(line[5])
		csv_row.Time 	= 	line[6]
		csv_row.Amount 	= 	line[7]
		csv_row.Use_Chip = 	line[8]
		csv_row.Merchant_Name 	= 	line[9]
		csv_row.Merchant_City 	= 	line[10]
		//Empty='missing_value'
		if line[11] == "" {
			csv_row.Merchant_State = "missing_value"
		}else{
			csv_row.Merchant_State = line[11]
		}
		//Empty=0.0
		if line[12] == "" {
			csv_row.Zip	= 0.0
		}else{
			csv_row.Zip,_ = strconv.ParseFloat(line[12], 64)
		}
		csv_row.MCC,_ 	= 	strconv.Atoi(line[13])
		//Empty='missing_value'
		if line[14] == "" {
			csv_row.Errors	= "missing_value"
		}else{
			csv_row.Errors = line[14]
		}
		//
		//fmt.Println(csv_row)
		//
		csv_row_json_byte, _ := json.Marshal(csv_row)
		//fmt.Println(csv_row_index)
		//fmt.Println(string(csv_row_json))
		//
		csv_row_dict[csv_row_index] = string(csv_row_json_byte)
		//
	}
	//
	fmt.Println(len(csv_row_dict))
	//
}

//Create Mapper funcation
//Mapper funcation used to map JSON string of test case into TFS required format
//Mapper implemented at the mapper.so
var Mapper_Gen_fun = Create_Mapper()
func Create_Mapper() func(string) string {
	//
	mapper, err := plugin.Open("mapper.so")
    if err != nil {
        panic(err)
    }
	
	//func ()
	Mapper_Init, err := mapper.Lookup("Mapper_Init")
	if err != nil {
		panic(err)
	}
	//
	//func Mapper_Gen(input_data_json string) string
	Mapper_Gen, err := mapper.Lookup("Mapper_Gen")
	if err != nil {
		panic(err)
	}
	//Mapper_Init_fun()
	Mapper_Init.(func())()
	//
	return Mapper_Gen.(func(string) string)
	//
}

//Create Redis Pool
//Redis be used to save the historical transactions data
//The historical transactions date preparation steps , please read readme 
var pool = newPool()
func newPool() *redis.Pool {
	//
	REDISADD 	= os.Getenv("REDISADD")
	fmt.Println("Redis URL=>",REDISADD)
	//
	redis_pool := &redis.Pool{
		MaxIdle: 80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", REDISADD)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
	//
	client_0 := redis_pool.Get()
	client_1 := redis_pool.Get()
	client_2 := redis_pool.Get()
	client_3 := redis_pool.Get()
	//
	client_0.Close()
	client_1.Close()
	client_2.Close()
	client_3.Close()
	//
	return redis_pool
}

//Call TFS get the predict results in JSON
//Input of TFS call is the output of mapper fucation
func Call_TFS(input_test_case_json string) Predict{
	//
	TFSADD	= os.Getenv("TFSADD")
	//url   := fmt.Sprintf("http://%v/v1/models/ccf_220_os_z_lstm:predict",TFSADD)
	url   := fmt.Sprintf("http://%v/v1/models/model:predict",TFSADD)
	//
    fmt.Println("URL:>", url)
	//fmt.Printf("Input JSON Str = %s \n", input_test_case_json)
    
	// var jsonStr = []byte(input_test_case_json)
	// post_req,_ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    // post_req.Header.Set("Content-Type", "application/json")
    // http_client := &http.Client{}
	//start := time.Now()
    // post_resp,post_err := http_client.Do(post_req)
    // if post_err != nil {
    //     panic(post_err)
    // }
	
    // defer post_resp.Body.Close()
	//
	client := &http.Client{}
    post_resp, err := client.Post(url,  "application/json", bytes.NewBuffer([]byte(input_test_case_json)))
    if err != nil {
        panic(err)
    }
    defer post_resp.Body.Close()
	//
    body,_ := ioutil.ReadAll(post_resp.Body)
	predict_str := string(body)
    //fmt.Print(predict_str,"\n")
	//
	var predict_ret Predict
	json.Unmarshal([]byte(predict_str), &predict_ret)
	fmt.Println(predict_ret)
	//
	return predict_ret
}


//Read Prarameter from Docker run command line
//Redis address and port number
var REDISADD string
//TFS address and port number
var TFSADD string


//The main entry of the REST API Server
func main() {
	//----------------------------------------
	// 1 Creat Test case Dict 
	//----------------------------------------
	//defer c1.Close()
	Create_test_case_dict(test_case_dict)

	//----------------------------------------
	// 2 Handle REST API Request 
	//----------------------------------------
	CCFHanlder := http.HandlerFunc(CCF)
	http.Handle("/ccf_inference", CCFHanlder)
	fmt.Println("Backend REST API Server started and ready for POST request") 
	fmt.Println("URL http://127.0.0.1:8080/ccf_inference")
	http.ListenAndServe(":8080", nil)
}

//Handle REST API Request
func CCF(w http.ResponseWriter, r *http.Request) {
	//----------------------------------------------------
	// 2.1 Read JSON parameter from REST API request
	//----------------------------------------------------
	headerContentTtype := r.Header.Get("Content-Type")
	if headerContentTtype != "application/json" {
		errorResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var e Tran
	// var unmarshalErr *json.UnmarshalTypeError
	decoder := json.NewDecoder(r.Body)
	//decoder.DisallowUnknownFields()
	err := decoder.Decode(&e)
	
	if err != nil {
		//
		errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		//
		return
	}
	fmt.Println("\nNew REST API Request =>",e.Tx_index)
	//errorResponse(w, "Success " + e.Tx_index, http.StatusOK)

	//----------------------------------------------------
	// 2.2 Get Test Case and mapping to the 7th transaction of TFS input
	//----------------------------------------------------
	tx_index := e.Tx_index
	start_Mapper  := time.Now()
	tx_json  := test_case_dict[tx_index]
	fmt.Println("Test Case Str B=",tx_index,tx_json)
	//
	if len(e.Tx_json) != 0 {
		tx_json = e.Tx_json
	}
	//
	fmt.Println("Test Case Str A=",tx_index,tx_json)
	mapped_json_str := Mapper_Gen_fun(tx_json)
	elapsed_Mapper := time.Since(start_Mapper)
	fmt.Println("\nMapper elapsed=",elapsed_Mapper)
	//fmt.Println("Mapped Str    =",mapped_json_str)


	//----------------------------------------------------
	// 2.3 Read 6 historical transactions fomr Redis in JSON string format
	//----------------------------------------------------
	client := pool.Get()
	defer client.Close()
	//
	start_Redis := time.Now()
	//
	redis_key_str := strconv.Itoa(tx_index)
	value, err 	:= client.Do("GET", strconv.Itoa(tx_index))
	redis_str  := string(value.([]byte))
	//fmt.Printf("Redis JSON Str = %s \n", redis_str)

	//Create TFS input
	//Combind 6 historical transacations with the new came in as input for TFS
	//fmt.Println( "redis_str_index_sixth:", redis_str_index_sixth, len(redis_str))
	tfs_input_str := redis_str[0 : len(redis_str)-2]
	tfs_input_str = tfs_input_str + ", [" + mapped_json_str + "]]}"
	//fmt.Println("tfs_input_str=",tfs_input_str)
	//
	elapsed_Redis := time.Since(start_Redis)
	fmt.Println("\nRedis elapsed=",elapsed_Redis)


	//Update the Redis with new transacation
	//Remove the first element of 6 historical transacations
	//Append the new transacation at the end of the string
	start_Redis_Update := time.Now()
	//
	redis_str_index_first := strings.Index( redis_str, "]], [[")
	redis_str_new_history := redis_str[redis_str_index_first + 5 : len(redis_str)-2]
	redis_str_new_history = `{"instances": [[` + redis_str_new_history + ", [" + mapped_json_str + "]]}"
	//
	//fmt.Println("tfs_input_str=",redis_str_new_history)
	//
	_,err = client.Do("Set", "W"+redis_key_str, redis_str_new_history)
	//
	elapsed_Redis_Update := time.Since(start_Redis_Update)
	fmt.Println("\nRedis Update elapsed=",elapsed_Redis_Update)
	//
	if err != nil {
		panic(err)
	}

	//----------------------------------------
    // 3 Call TFS REST API with the JSON
    //----------------------------------------
	start_TFS := time.Now()
	predict_ret := Call_TFS(tfs_input_str)
	elapsed_TFS := time.Since(start_TFS)
	fmt.Println("\nTFS elapsed=",elapsed_TFS)
	//
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := make(map[string]string)
	//
	resp["TX         Index"] = strconv.Itoa(e.Tx_index)
	resp["Predict    Value"] = strconv.FormatFloat(predict_ret.Predictions[6][0][0], 'E', -1, 64)
	resp["Elapsed   Mapper"] = fmt.Sprintf("%v",elapsed_Mapper)
	resp["Elapsed  Redis R"] = fmt.Sprintf("%v",elapsed_Redis)
	resp["Elapsed  Redis U"] = fmt.Sprintf("%v",elapsed_Redis_Update)
	resp["Elapsed      TFS"] = fmt.Sprintf("%v",elapsed_TFS)
	//
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
	//
	return
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}
