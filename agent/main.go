package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

var (
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

func evalWithDelay(expr string, timings []string) (result float64) { // функция, необходимая для работы solve, ждёт столько, сколько передали, потом решает выражение
	intExecutionTimings := []int{}
	for i := 0; i < len(timings); i++ {
		val, _ := strconv.Atoi(timings[i])
		intExecutionTimings = append(intExecutionTimings, val)
	}
	ToExecutePlus, ToExecuteMinus, ToExecuteMultiply, ToExecuteDivide := intExecutionTimings[0], intExecutionTimings[1], intExecutionTimings[2], intExecutionTimings[3]
	time.Sleep(time.Second*time.Duration(ToExecuteMinus) + time.Second*time.Duration(ToExecutePlus) + time.Second*time.Duration(ToExecuteMultiply) + time.Second*time.Duration(ToExecuteDivide))
	return eval(expr)
}

func sendToOrchestraByGet(res float64) { // функция, необходимая для работы solve, отправляет решённое выражение
	addr := fmt.Sprintf("http://127.0.0.1:%s/receiveresult/?Result=%.3f&Id=%s&AgentPort=8999", ConnectedTo, Result, Id)
	fmt.Println(addr)
	//addr := "http://127.0.0.1:8080/receiveresult/?Result=15.5&Id=1488"
	_, _ = http.Get(addr)
}

func Connect(w http.ResponseWriter, r *http.Request) { // /connect/ Даёт порт, на котором хостится оркестр
	ConnectedTo = r.URL.Query().Get("HostPort")
	fmt.Println(ConnectedTo)

}

func Solve(w http.ResponseWriter, r *http.Request) { // /solve/ получает выражение и онправляет его оркустру
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
	fmt.Println("hb received", ConnectedTo)
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "you shouldnt be here)")
	})
	http.HandleFunc("/connect/", Connect)
	http.HandleFunc("/solve/", Solve)
	http.HandleFunc("/heartbeat/", HandleHeratbeat)

	http.ListenAndServe(":8999", nil)
}

// 127.0.0.1:8999/connect/?HostPort=8080
// 127.0.0.1:8999/solve/?Expression=(2%2B2*5-3)%2F2&Id=1488&ExecutionTimings=1!2!3!4
// + == %2B
// / == %2F

// CHECK LESSON 2 MAIN GO
