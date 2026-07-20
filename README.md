# policy-playground

## Описание

policy-playground — CLI-приложение, представляющее собой стенд проверки политик безопасности на сценариях действий пользователя.

На вход программа получает сценарий и список политик. Далее из сценария берутся события и каждое проверяется по условиям каждой политики. Информация о сработавших политиках для проверяемых событий и причины их срабатывания записываются в отдельный JSONL-файл.

## Возможности

- Чтение сценария из YAML-файлов.
- Чтение политик из YAML-файлов.
- `equals`: проверка равенства значения поля события значению, указанному в условии политики.
- `contains`: проверяет, содержится ли значение, указанное в условии политики, в массиве значений поля события.
- `in`: проверка вхождения значения поля события в список допустимых значений, указанный в политике.
- `all`: группа условий, которая срабатывает только в том случае, если успешно выполнены все входящие в неё условия.
- Запись срабатываний политик и причин каждого срабатывания в формате JSONL. Каждое срабатывание записывается отдельной строкой.

## Требования

- Go 1.26.4 или совместимая версия
- GNU Make — для использования команд `make test` и `make demo`

## Сборка

Использовать из корня проекта команду: `go build -o policy-playground.exe ./cmd/policy-playground`. Это нужно, чтобы получить исполняемый файл.

Команда создаёт исполняемый файл policy-playground.exe в корне проекта.

## Запуск

Общий синтаксис команды: `policy-playground run --scenario <путь> --policies <путь> --out <путь>`

## Флаги команды run

- `--scenario`: путь к файлу со сценарием.
- `--policies`: путь к файлу с политиками.
- `--out`: путь к выходному JSONL-файлу со срабатываниями; по умолчанию alerts.jsonl.

Способы использования:

1) После сборки: запустить готовый исполняемый файл

Для запуска использовать команду из корня проекта: `./policy-playground.exe run --scenario testdata/control/scenario.yaml --policies testdata/control/policies.yaml --out testdata/control/alerts.jsonl`

2) Без предварительной сборки

Для запуска использовать команду из корня проекта: `go run ./cmd/policy-playground run --scenario testdata/control/scenario.yaml --policies testdata/control/policies.yaml --out testdata/control/alerts.jsonl`

## Формат входных данных

### Сценарий

YAML-файл со сценарием содержит следующие данные о нем:

- `scenario_id`: ID сценария
- `name`: название
- `users`: пользователи
- `events`: события

Каждое событие содержит данные действия пользователя:

- `event_id`: ID события
- `time`: время события
- `user_id`: ID пользователя
- `action`: действие
- `object_type`: тип объекта (`file`, `message`, `archive`, `record`)
- `file_name`: название файла
- `file_ext`: расширение файла
- `content_classes`: классы контента
- `channel`: канал передачи данных (`local`, `email`, `usb`, `printer`, `cloud`, `messenger`)
- `destination_type`: тип назначения (`none`, `internal`, `external`, `usb`, `printer`, `cloud`)
- `size_bytes`: размер

При этом следующие поля являются необязательными:

- `file_name`: название файла
- `file_ext`: расширение файла
- `content_classes`: классы контента
- `size_bytes`: размер

Все остальные поля события являются обязательными.

Каждый пользователь имеет следующие данные:

- `user_id`: ID пользователя
- `department`: отдел пользователя
- `role`: роль пользователя

Пример сценария:

```yaml
scenario_id: scenario_001
name: External client data send
users:
- user_id: user_001
  department: sales
  role: manager
events:
- event_id: evt_001
  time: "10:00"
  user_id: user_001
  action: open_file
  object_type: file
  file_name: client_base.xlsx
  file_ext: xlsx
  content_classes: [client_data, personal_data]
  channel: local
  destination_type: none
  size_bytes: 204800
- event_id: evt_002
  time: "10:05"
  user_id: user_001
  action: email_send
  object_type: file
  file_name: client_base.xlsx
  file_ext: xlsx
  content_classes: [client_data, personal_data]
  channel: email
  destination_type: external
  size_bytes: 204800
```

### Политики

Корневое поле `policies` содержит список политик.

Каждая политика в свою очередь содержит следующие данные:

- `policy_id`: ID политики
- `name`: название
- `severity`: важность (`low`, `medium`, `high`, `critical`)
- `description`: описание
- `condition`: условие

Поле `description` является необязательным. Остальные перечисленные поля обязательны.

В текущей реализации поле `condition` поддерживает операторы `equals`, `in`, `contains` и логическую группу `all`.

Пример файла политик:

```yaml
policies:
- policy_id: pol_external_client_data
  name: Client data to external channel
  severity: high
  description: Detects sending client data to external destination
  condition:
    all:
    - field: action
      equals: email_send
    - field: destination_type
      equals: external
    - field: content_classes
      contains: client_data
```

## Формат выходных данных

Срабатывания политик сохраняются в JSONL-файл, путь к которому указывается через `--out`, и каждое срабатывание записывается отдельной строкой.

Поля одной записи:

- `policy_id`: ID сработавшей политики
- `policy_name`: название политики
- `severity`: уровень важности
- `event_id`: ID события, на котором сработала политика
- `user_id`: ID пользователя события
- `matched`: признак срабатывания
- `reasons`: список причин срабатывания

Пример одного срабатывания:

```json
{"policy_id":"pol_external_client_data","policy_name":"Client data to external channel","severity":"high","event_id":"evt_002","user_id":"user_001","matched":true,"reasons":["action equals email_send","destination_type equals external","content_classes contains client_data"]}
```

## Тестирование

Тесты запускаются из корня проекта.
Для запуска используется команда `make test`, которая запускает все тесты проекта.

## Контрольный запуск

Для демонстрации используется набор данных из папки `testdata/control`.

Для запуска используется команда `make demo`. Она использует:
- `testdata/control/scenario.yaml`
- `testdata/control/policies.yaml`

Результат записывается в `testdata/control/alerts.jsonl`

В результате контрольного запуска было получено одно срабатывание.
Сработала политика `pol_external_client_data` для события `evt_002`.

### Характеристики машины

- ОС: Windows 11
- Процессор: AMD Ryzen 5 5500U with Radeon Graphics
- ОЗУ: 8 ГБ

## Алгоритм работы

1. Программа получает пути к `scenario.yaml`, `policies.yaml` и выходному файлу.
2. Загружает сценарий и политики.
3. Последовательно перебирает события сценария.
4. Для каждого события проверяет каждую политику.
5. Проверка политики выполняется через её `condition`:
    - простые операторы `equals`, `in`, `contains`.
    - либо группа `all`, где должны выполниться все вложенные условия.
6. Если политика сработала, формируется срабатывание с информацией о политике, событии и причинах срабатывания.
7. Срабатывание сразу записывается в JSONL.
8. Если политика не сработала, запись для неё не создаётся.

## Известные ограничения

На данный момент не реализованы:

- Группа `any`
- `exists`
- Команда `explain`
- Markdown-отчеты
- Режим "почему не сработало"
- Оконные условия
- HTML-отчет
- Benchmark на 100 000 событий и 100 политик