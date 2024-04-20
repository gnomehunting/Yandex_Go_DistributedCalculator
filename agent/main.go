package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

var (
	AgentPort   string
	HostPort    string
	ConnectedTo string
	Expression  string
	Id          string
	Result      float64
)

func eval(tosolve string) float64 { // костыль, чтобы функция eval, импортированная из библиотеки, работала так, как надо мне
	expression, _ := govaluate.NewEvaluableExpression(tosolve)
	parameters := make(map[string]interface{}, 8)
	result, _ := expression.Evaluate(parameters)
	toret := result.(float64)
	return float64(toret)
}

func evalWithDelay(expr string, timings []string) (result float64) { // функция, необходимая для работы solve, ждёт заданные оркестром тайминги * вхождение знаков в выражение, потом решает выражение
	intExecutionTimings := []int{}
	for i := 0; i < len(timings); i++ {
		val, _ := strconv.Atoi(timings[i])
		intExecutionTimings = append(intExecutionTimings, val)
	}
	ToExecutePlus, ToExecuteMinus, ToExecuteMultiply, ToExecuteDivide := strings.Count(expr, "+")*intExecutionTimings[0], strings.Count(expr, "-")*intExecutionTimings[1], strings.Count(expr, "*")*intExecutionTimings[2], strings.Count(expr, "/")*intExecutionTimings[3]
	time.Sleep(time.Second*time.Duration(ToExecuteMinus) + time.Second*time.Duration(ToExecutePlus) + time.Second*time.Duration(ToExecuteMultiply) + time.Second*time.Duration(ToExecuteDivide))
	return eval(expr)
}

func sendToOrchestraByGet(res float64) { // функция, необходимая для работы solve, отправляет решённое выражение орекстратору
	addr := fmt.Sprintf("http://127.0.0.1:%s/receiveresult/?Result=%.3f&Id=%s&AgentPort=%s", ConnectedTo, Result, Id, AgentPort)
	fmt.Println(addr)
	_, _ = http.Get(addr)
}

func Connect(w http.ResponseWriter, r *http.Request) { // /connect/ орекстр делает гет запрос сюда, чтобы дать агенту знать о порте оркестра
	ConnectedTo = r.URL.Query().Get("HostPort")
	fmt.Println(ConnectedTo)

}

func Solve(w http.ResponseWriter, r *http.Request) { // /solve/ получает, решает и отправляет выражение оркестру, гетами
	Expression = r.URL.Query().Get("Expression")
	Id = r.URL.Query().Get("Id")
	ExecutionTimings := strings.Split(r.URL.Query().Get("ExecutionTimings"), "!")
	go func() {
		Result = evalWithDelay(Expression, ExecutionTimings)
		sendToOrchestraByGet(Result)
		fmt.Println("finished solving")
	}()
	fmt.Fprintln(w, "Started Solving")

}

func HandleHeratbeat(w http.ResponseWriter, r *http.Request) { // /heartbeat/ оркестр отправляет сюда хартбиты
	ConnectedTo = r.URL.Query().Get("HostPort")
	fmt.Println("hb received", ConnectedTo)
}

func main() {
	AgentPort = os.Args[1] // через os.args задаётся порт, на котором будет работать агент
	if AgentPort == "" {
		log.Fatal("PORT not set")
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "you shouldnt be here)")
	})
	http.HandleFunc("/connect/", Connect)
	http.HandleFunc("/solve/", Solve)
	http.HandleFunc("/heartbeat/", HandleHeratbeat)

	http.ListenAndServe(":"+AgentPort, nil) //обратботка эндпоинтов
}
