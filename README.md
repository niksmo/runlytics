# Сервис сбора метрик и алертинга «Runlytics»

## Функционал

Агент собирает метрики из пакета `runtime` + `RandomValue и PollCount` и отправляет их на сервер по протоколу HTTP.

Сервер сохраняет полученные метрики от агента в памяти и файле или в базе данных.

## Агент

Собирает метрики двух типов:
- `gauge` - тип `float64`
- `counter` - тип `int64`

### Список метрик типа `gauge`

- `RandomValue` — обновляемое произвольное значение
- `Alloc`
- `BuckHashSys`
- `Frees`
- `GCCPUFraction`
- `GCSys`
- `HeapAlloc`
- `HeapIdle`
- `HeapInuse`
- `HeapObjects`
- `HeapReleased`
- `HeapSys`
- `LastGC`
- `Lookups`
- `MCacheInuse`
- `MCacheSys`
- `MSpanInuse`
- `MSpanSys`
- `Mallocs`
- `NextGC`
- `NumForcedGC`
- `NumGC`
- `OtherSys`
- `PauseTotalNs`
- `StackInuse`
- `StackSys`
- `Sys`
- `TotalAlloc`

### Список метрик типа `counter`

- `PollCount` — счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета `runtime`

### Конфигурирование Агента

Агент поддерживает конфигурирование следующими флагами и переменными:

- URL адрес сервера сбора метрик: переменная окружения `ADDRESS` или флаг `-a` (по умолчанию `http://localhost:8080`)
- ключ для хэширования запроса: переменная окружения `KEY` или флаг `-k` (по умолчанию не задан)
- уровень логирования: переменная окружения `LOG_LVL` или флаг `-log` (по умолчанию `info`)
- интервал сбора метрик в секундах: переменная окружения `POLL_INTERVAL` или флаг `-p` (по умолчанию `2`)
- интервал отправки метрик на сервер в секундах: переменная окружения `REPORT_INTERVAL` или флаг `-r` (по умолчанию `10`)
- количество воркер отправки метрик на сервер: переменная окружения `RATE_LIMIT` или флаг `-l` (по умолчанию `1`)


## Сервер

### Сводное HTTP API Сервера

* `GET /` - получение списка метрик в виде html страницы
* `GET /ping` - хелсчек
* `GET /value/{type}/{name}` - получнеие значения метрики по типу и назнванию
* `POST /value/` - получение значения метрики по типу и названию, запрос в формате `json`
* `POST /update/{type}/{name}/{value}` - запись/обновление метрики
* `POST /update/` - запись/обновление метрики, запрос в формате `json`
* `POST /updates/` - запись/обновление списка метрик, запрос в формате `json`

### Конфигурирование Сервера

Сервер поддерживает конфигурирование следующими флагами и переменными:

- адрес и порт прослушиваемого сервером: переменная окружения `ADDRESS` или флаг `-a` (по умолчанию `localhost:8080`)
- ключ хэширования для проверки запроса от агента: переменная окружения `KEY` или флаг `-k` (по умолчанию не задан)
- уровень логирования: переменная окружения `LOG_LVL` или флаг `-log` (по умолчанию `info`)
- сохранение метрик в памяти и файле:
    - путь к файлу: переменная окружения `FILE_STORAGE_PATH` или флаг `-f` (по умолчанию `{project_dir}\storage.json`)
    - интервал (в секундах) сохранения метрик из памяти в файл: переменная окружения `STORE_INTERVAL` или флаг `-i` (по умолчанию `300`)
    - признак восстановления метрик из файла в память при запуске сервера: переменная окружения `RESTORE` или флаг `-r` (по умолчанию `1`)
- сохранение метрик в базе данных:
    - адрес подключения к базе данных: переменная окружения `DATABASE_DSN` или флаг `-d` (по умолчанию не задан)
