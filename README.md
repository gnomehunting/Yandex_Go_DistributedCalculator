0. >Необходимые требования:
    Существует Readme документ, в котором описано, как запустить систему и как ей пользоваться.
    Это может быть docker-compose, makefile, подробная инструкция - на ваш вкус (этот readme)
    Если вы предоставляете только http-api, то
    в Readme описаны примеры запросов с помощью curl-a или любым дргуми понятными образом
    примеры полны и понятно как их запустить
    Этот пункт дает 10 баллов. Без наличия такого файла - решение не проверяется.
    - Сделано
1. >Программа запускается и все примеры с вычислением арифметических выражений корректно работают - 10 баллов
    - По идее - да
2. >Программа запускается и выполняются произвольные примеры с вычислением арифметических выражений - 10 баллов
    - По идее - да
3. >Можно перезапустить любой компонент системы и система корректно обработает перезапуск (результаты сохранены, система продолжает работать) - 10 баллов
    - Можно перезапустить агент, если сделать это до истечени времени, кторое он должен показываться, он "воскреснет" если он уже "умер" - создастя новая структура агент на том же порте, "умершего" видно не будет на странице, оркестратор перезапускать нельзя.
4. >Система предосталяет графический интерфейс для вычисления арифметических выражений - 10 баллов
    - localhost:<ваш порт>/calculator/
5. >Реализован мониторинг воркеров - 20 баллов
    - localhost:<ваш порт>/agents/
6. >Реализован интерфейс для мориторинга воркеров - 10 баллов
    - localhost:<ваш порт>/agents/
7. >Вам понятна кодовая база и структура проекта - 10 баллов (это субъективный критерий, но чем проще ваше решение - тем лучше).
    Проверяющий в этом пункте честно отвечает на вопрос: "Смогу я сделать пулл-реквест в проект без нервного срыва"
    - Я постарался добавить достаточно комментариев
8. >У системы есть документация со схемами, которая наглядно отвечает на вопрос: "Как это все работает" - 10 баллов
    - [схема](https://github.com/demonShaco69/Yandex_Go_DistributedCalculator/blob/main/scheme%20of%20preject.png) + рядом с каждой функций написано, что она делает
9. >Выражение должно иметь возможность выполняться разными агентами - 10 баллов
    - Выражения раскидываются по разным агентам, вычисляются параллельно, но не делятся на субвыражения


- КАК ЭТИМ ПОЛЬЗОВАТЬСЯ:
    есть 2 main.go в папке orchestra и в папке agent, чтобы программа заработала, запускаете оба файла на разных портах, переходите на начальную страницу оркестратора (можно запустить сколько угодно агентов, но нужно на разных портах, порт передаётся агенту через os.args) 
- ЗАПУСК АГЕНТА go run .\agent\main.go <ваш порт>
- агенту необходима библиотека "github.com/Knetic/govaluate", если что-то пойдёт не так - ее можно импортировать:  go get github.com/Knetic/govaluate (запустить из терминала из папки orchestra)
- ЗАПУСК ОРКЕСТРА go run .\orchestra\main.go <ваш порт>


- КАК ПОЛЬЗОВАТЬСЯ ВЕБ СТРАНИЦАМИ:
    Чтобы попасть на веб страницу, нужно перейти на localhost:<порт орекстра>
    3 верхних заголовка - гиперссылки, позволяют перемещаться по страницам
- КАЛЬКУЛЯТОР:
    есть поле, в которое можно ввести выражение, вводить можно цифры, знаки(-+/*), и скобки, "=" вводить нельзя, пробелы вводить нельзя, повторяющееся уравнение не будет отправлено на сервер, если будет отправленно некорректное выражение, ему будет присвоен статус invalid
    выражение отправляется на сервер кнопкой Solve
    можно обновить страницу кнопкой Refresh
    на странице имеется список выражений,(текст выражения/результат/id/статус)
- ТАЙМИНГИ:
    показаны тайминги на выполнение действий и время показа на странице Agents не отвечающих серверов
    5 полей, позволяют изменять вышеназванные значения, кнопка submit отправляет их
    можжно вводить только цифры, иначе тайминг не изменится
- АГЕНТЫ:
    кнопка refresh
    поле ввода, в которое нужно ввести порт запущенного вами ранее агента, поле add
    есть список агентов(порт, статус, кол-во раз агент не принял хартбит)

!!!Если у вас возник вопрос, напишите мне в тг: https://t.me/xdd42 !!!
