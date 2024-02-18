package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Agent struct {
	Status       string
	Port         string
	NotResponded int
	Display      bool
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

func isValidExpression(expression string) bool {
	re := regexp.MustCompile(`^\d+([\+\-\*\/]\d+)+$`)
	withoutcommas := expression
	withoutcommas = strings.ReplaceAll(withoutcommas, "(", "")
	withoutcommas = strings.ReplaceAll(withoutcommas, ")", "")

	ismatching := re.MatchString(withoutcommas)

	stack := []rune{}

	for _, char := range expression {
		if char == '(' {
			stack = append(stack, '(')
		} else if char == ')' {
			if len(stack) == 0 {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0 && ismatching
}

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
	needtoadd := true
	for i := range MapOfExpressions {
		if MapOfExpressions[i].Text == txt {
			needtoadd = false
		}
	}
	if needtoadd {
		if isValidExpression(txt) {
			MapOfExpressions[len(MapOfExpressions)] = Expression{Text: txt, Id: strconv.Itoa(len(MapOfExpressions)), Result: "0", Status: "unsolved"}
		} else {
			MapOfExpressions[len(MapOfExpressions)] = Expression{Text: txt, Id: strconv.Itoa(len(MapOfExpressions)), Result: "0", Status: "invalid"}
		}
	}

	http.Redirect(w, r, "/calculator/", http.StatusSeeOther)
}

func CalculatorPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("orchestra/calculator.html"))
	tmpl.Execute(w, MapOfExpressions)
}

func ChangeTimings(w http.ResponseWriter, r *http.Request) {
	_, err1 := strconv.Atoi(r.FormValue("plu"))
	_, err2 := strconv.Atoi(r.FormValue("min"))
	_, err3 := strconv.Atoi(r.FormValue("mul"))
	_, err4 := strconv.Atoi(r.FormValue("div"))
	_, err5 := strconv.Atoi(r.FormValue("whb"))
	if err1 == nil {
		newTimings.Plus = r.FormValue("plu")
	}
	if err2 == nil {
		newTimings.Minus = r.FormValue("min")
	}
	if err3 == nil {
		newTimings.Multiply = r.FormValue("mul")
	}
	if err4 == nil {
		newTimings.Divide = r.FormValue("div")
	}
	if err5 == nil {
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
	_, err := strconv.Atoi(port)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/agents/", http.StatusSeeOther)
	} else {
		addr := fmt.Sprintf("http://127.0.0.1:%s/connect/?HostPort=%s", port, OrchestraPort)
		_, _ = http.Get(addr)
		ListOfAgents = append(ListOfAgents, Agent{Port: port, Status: "notresponding", NotResponded: 0, Display: true})
		http.Redirect(w, r, "/agents/", http.StatusSeeOther)
	}

}
func AgentsPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("orchestra/agents.html"))
	tmpl.Execute(w, ListOfAgents)
}

func heartbeat() {
	for {
		if len(ListOfAgents) != 0 {
			for i, agent := range ListOfAgents {
				if ListOfAgents[i].NotResponded >= 1 {
					ListOfAgents[i].Status = "notresponding"
				}
				if ListOfAgents[i].NotResponded >= 5 {
					ListOfAgents[i].Status = "dead"
				}
				if ListOfAgents[i].Status != "dead" {
					heartbeataddr := fmt.Sprintf("http://127.0.0.1:%s/heartbeat/?HostPort=%s", agent.Port, OrchestraPort)
					_, err := http.Get(heartbeataddr)
					if err != nil {
						ListOfAgents[i].NotResponded++
						continue
					} else {
						if ListOfAgents[i].Status != "busy" {
							ListOfAgents[i].NotResponded = 0
							ListOfAgents[i].Status = "online"
						}
					}
				} else {
					ListOfAgents[i].Display = false
				}

			}
			ttw, _ := strconv.Atoi(newTimings.DisplayTime)
			time.Sleep(duration(float64(ttw / 5)))
		}
	}

}

func duration(f float64) time.Duration {
	return time.Duration(f * 1e9)
}

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
							_, err := http.Get(addr)
							if err != nil {
								fmt.Println(err)
							} else {
								MapOfExpressions[i] = Expression{Text: MapOfExpressions[i].Text, Id: MapOfExpressions[i].Id, Result: MapOfExpressions[i].Result, Status: "solving"}
								ListOfAgents[j].Status = "busy"
							}

						}
					}
				}
			}
		}
	}

}

func main() {
	OrchestraPort = os.Args[1]
	fmt.Println(OrchestraPort)
	if OrchestraPort == "" {
		log.Fatal("PORT not set")
	}
	newTimings.Plus = "1"
	newTimings.Minus = "1"
	newTimings.Multiply = "1"
	newTimings.Divide = "1"
	newTimings.DisplayTime = "20"

	MapOfExpressions = make(map[int]Expression)
	//MapOfExpressions[0] = Expression{Text: "2-2*2", Id: "0", Result: "0", Status: "unsolved"}
	//MapOfExpressions[1] = Expression{Text: "5*8-6", Id: "1", Result: "0", Status: "unsolved"}
	//MapOfExpressions[2] = Expression{Text: "66-99*2", Id: "2", Result: "0", Status: "unsolved"}

	//newAgent1 := Agent{Status: "online", Port: "8999"}
	//newAgent2 := Agent{Status: "online", Port: "8998"}
	//ListOfAgents = append(ListOfAgents, newAgent1)
	//ListOfAgents = append(ListOfAgents, newAgent2)

	go heartbeat()
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
