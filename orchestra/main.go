package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Agent struct {
	Status string
	Port   string
}
type Expression struct {
	Text   string
	Id     string
	Result string
	Status string
}
type Timings struct {
	Plus        string
	Minus       string
	Multiply    string
	Divide      string
	DisplayTime string
}

var OrchestraPort string
var MapOfExpressions map[int]Expression
var ListOfAgents []Agent
var IdCounter int
var newTimings Timings

func ReceiveResult(w http.ResponseWriter, r *http.Request) { // /receiveresult/ агент отправляет сюда решённое выражение
	result := r.URL.Query().Get("Result")
	id := r.URL.Query().Get("Id")
	port := r.URL.Query().Get("AgentPort")
	intid, _ := strconv.Atoi(id)
	fmt.Println(result, id)
	MapOfExpressions[intid] = Expression{Text: MapOfExpressions[intid].Text, Id: MapOfExpressions[intid].Id, Status: "solved", Result: result}
	for i, agent := range ListOfAgents {
		if agent.Port == port {
			ListOfAgents[i].Status = "online"
		}
	}
}

func AddExpression(w http.ResponseWriter, r *http.Request) {
	txt := r.FormValue("item")
	MapOfExpressions[len(MapOfExpressions)] = Expression{Text: txt, Id: strconv.Itoa(len(MapOfExpressions)), Result: "0", Status: "unsolved"}
	http.Redirect(w, r, "/calculator/", http.StatusSeeOther)
}
func CalculatorPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("orchestra/calculator.html"))
	tmpl.Execute(w, MapOfExpressions)
}

func ChangeTimings(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.FormValue("plus"))
	if r.FormValue("plu") != "" {
		newTimings.Plus = r.FormValue("plu")
	}
	if r.FormValue("min") != "" {
		newTimings.Minus = r.FormValue("min")
	}
	if r.FormValue("mul") != "" {
		newTimings.Multiply = r.FormValue("mul")
	}
	if r.FormValue("div") != "" {
		newTimings.Divide = r.FormValue("div")
	}
	if r.FormValue("whb") != "" {
		newTimings.DisplayTime = r.FormValue("whb")
	}
	http.Redirect(w, r, "/timings/", http.StatusSeeOther)
}

func TimingsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("orchestra/timings.html"))
	tmpl.Execute(w, newTimings)
}

func AddAgent(w http.ResponseWriter, r *http.Request) {
	port := r.FormValue("agentport")
	addr := fmt.Sprintf("http://127.0.0.1:%s/connect/?HostPort=%s", port, OrchestraPort)
	_, _ = http.Get(addr)
	ListOfAgents = append(ListOfAgents, Agent{Port: port, Status: "online"})
	fmt.Println(ListOfAgents)
	http.Redirect(w, r, "/agents/", http.StatusSeeOther)
}
func AgentsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("orchestra/agents.html"))
	tmpl.Execute(w, ListOfAgents)
}

/*
	func StartHeartbeat(agent *Agent) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { //HEARTBEAT сюда передаётся порт агента, оркест начинает слдать ему хартбиты
			heartbeataddr := fmt.Sprintf("http://127.0.0.1:%s/heartbeat/?HostPort=%s", agent.Port, OrchestraPort)
			go func() {
				for {
					time.Sleep(5 * time.Second)
					_, err := http.Get(heartbeataddr)
					if err != nil {
						fmt.Println("Error sending heartbeat :", err)
						continue
					}
					fmt.Println("Heartbeat sent to server")
				}
			}()
		}
	}
*/
func mainSolver() {
	for {
		time.Sleep(time.Second)
		if len(MapOfExpressions) != 0 && len(ListOfAgents) != 0 {
			for i := 0; i < len(MapOfExpressions); i++ {
				if MapOfExpressions[i].Status == "unsolved" {
					for j := range ListOfAgents {
						if ListOfAgents[j].Status == "online" && MapOfExpressions[i].Status == "unsolved" {
							textwithreplacements := MapOfExpressions[i].Text
							textwithreplacements = strings.ReplaceAll(textwithreplacements, "+", "%2B")
							textwithreplacements = strings.ReplaceAll(textwithreplacements, "/", "%2F")
							addr := fmt.Sprintf("http://127.0.0.1:%s/solve/?Expression=%s&Id=%s&ExecutionTimings=%s!%s!%s!%s", ListOfAgents[j].Port, textwithreplacements, MapOfExpressions[i].Id, newTimings.Plus, newTimings.Minus, newTimings.Multiply, newTimings.Divide)
							fmt.Println(addr)
							_, err := http.Get(addr)
							if err != nil {
								fmt.Println(err)
							} else {
								MapOfExpressions[i] = Expression{Text: MapOfExpressions[i].Text, Id: MapOfExpressions[i].Id, Result: MapOfExpressions[i].Result, Status: "solving"}
								ListOfAgents[j].Status = "busy"
								fmt.Println(ListOfAgents)
							}

						}
					}
				}
			}
		}
	}

}

func main() {
	newTimings.Plus = "1"
	newTimings.Minus = "1"
	newTimings.Multiply = "1"
	newTimings.Divide = "1"
	newTimings.DisplayTime = "1"

	MapOfExpressions = make(map[int]Expression)
	//MapOfExpressions[0] = Expression{Text: "2-2*2", Id: "0", Result: "0", Status: "unsolved"}
	//MapOfExpressions[1] = Expression{Text: "5*8-6", Id: "1", Result: "0", Status: "unsolved"}
	//MapOfExpressions[2] = Expression{Text: "66-99*2", Id: "2", Result: "0", Status: "unsolved"}

	//newAgent1 := Agent{Status: "online", Port: "8999"}
	//newAgent2 := Agent{Status: "online", Port: "8998"}
	//ListOfAgents = append(ListOfAgents, newAgent1)
	//ListOfAgents = append(ListOfAgents, newAgent2)

	OrchestraPort = "8080"
	go mainSolver()

	http.HandleFunc("/receiveresult/", ReceiveResult)
	http.HandleFunc("/calculator/", CalculatorPage)
	http.HandleFunc("/timings/", TimingsPage)
	http.HandleFunc("/agents/", AgentsPage)
	http.HandleFunc("/add/", AddExpression)
	http.HandleFunc("/changetimings/", ChangeTimings)
	http.HandleFunc("/addagent/", AddAgent)
	http.ListenAndServe(":"+OrchestraPort, nil)
}
