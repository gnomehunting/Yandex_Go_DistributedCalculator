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

type Agent struct { // структура, которая отражает агента: поле статус(notresponding/busy/dead/online), порт, поле отражающее кол во пропущенных хартбитов, поле, которое говорит - показывать агента на вебстранице, или нет
	Status       string
	Port         string
	NotResponded int
	Display      bool
}
type Expression struct { // структура, которая отражает выражение: поле текст(выражение записанное в строку, подлежит eval'у агентом), id, результат, и статус(notsolved/solving/solved/invalid)
	Text   string
	Id     string
	Result string
	Status string
}
type Timings struct { // время в секундах, требуемое для операции и время показа сервера, не принимающего хартибиты(задаётся на 2 веб странице сервера)
	Plus        string
	Minus       string
	Multiply    string
	Divide      string
	DisplayTime string
}

var OrchestraPort string                //порт оркестратора
var MapOfExpressions map[int]Expression // мапа со структурами Expression key == Expression.id
var ListOfAgents []Agent                // список агентов
var newTimings Timings                  // тайминги

func isValidExpression(expression string) bool { // функция, которая проверяет выражение на правильность (скобки/знаки/цифры)
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func ReceiveResult(w http.ResponseWriter, r *http.Request) { // /receiveresult/ агент отправляет выражение на эндпоинт /receiveresult/ и оно изменяется в мапе MapOfEspressions, Агенту, решившему и отправившему результат присваивается статус online
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func AddExpression(w http.ResponseWriter, r *http.Request) { // /add/ добавляет выражение к списку ListOfExpressions с помощью формы на странице калькулятор, попутно проверяя его с помощью вышеописанной функции isValidExpression
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func CalculatorPage(w http.ResponseWriter, r *http.Request) { // /calculator/ отрисовка страницы калькулятор, в темплейт передаётся мапа выражений
	tmpl := template.Must(template.ParseFiles("orchestra/calculator.html"))
	tmpl.Execute(w, MapOfExpressions)
}

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func ChangeTimings(w http.ResponseWriter, r *http.Request) { // /changetimings/ меняет вышеописанные тайминги при помощи форм на странице timings
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func TimingsPage(w http.ResponseWriter, r *http.Request) { // /timings/ отрисовка страницы с таймингами, в темплейт передаются тайминги
	tmpl := template.Must(template.ParseFiles("orchestra/timings.html"))
	tmpl.Execute(w, newTimings)
}

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func AddAgent(w http.ResponseWriter, r *http.Request) { // /addagent/ функция добавления агента через форму на страницу агент мониторинга, проверяет введённый порт на правильность, отправляет агенту пинг, чтобы он знал о существовании агента
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func AgentsPage(w http.ResponseWriter, r *http.Request) { // /agents/ отрисовка страницы с агентами, в темплейт передаётся список агентов
	tmpl := template.Must(template.ParseFiles("orchestra/agents.html"))
	tmpl.Execute(w, ListOfAgents)
}

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func heartbeat() { // запускается параллельно, отправляет хартбит всем подключенным агентам, если агент пропускает хартбит, ему даётся статус notresponding, если пропускает 5 - статус dead + агент перестаёт показываться на странице мониторинга
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

func duration(f float64) time.Duration { // функция, которая нужна для удобства работы с time.Duration и флоат64
	return time.Duration(f * 1e9)
}

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func mainSolver() { // функция, которая непрерыванол пробегается по списку агентов и мапе выражений чтобы выдать свободным агентам выражение
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

// ПЕРЕДЕЛАТЬ, ВМЕСТО МАПЫ БАЗУ ДАННЫХ
func main() {
	OrchestraPort = os.Args[1] // через os.args задаётся порт, на котором будет работать оркестратор
	fmt.Println(OrchestraPort)
	if OrchestraPort == "" {
		log.Fatal("PORT not set")
	}

	newTimings.Plus = "1"
	newTimings.Minus = "1"
	newTimings.Multiply = "1"
	newTimings.Divide = "1"
	newTimings.DisplayTime = "20" //дефолтные тайминги

	MapOfExpressions = make(map[int]Expression) // make map чтобы она не была nil мапой

	go heartbeat()
	go mainSolver()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/calculator/", http.StatusSeeOther)
	})
	http.HandleFunc("/receiveresult/", ReceiveResult)
	http.HandleFunc("/calculator/", CalculatorPage)
	http.HandleFunc("/timings/", TimingsPage)
	http.HandleFunc("/agents/", AgentsPage)
	http.HandleFunc("/add/", AddExpression)
	http.HandleFunc("/changetimings/", ChangeTimings)
	http.HandleFunc("/addagent/", AddAgent)
	http.ListenAndServe(":"+OrchestraPort, nil) //обратботка эндпоинтов
}
