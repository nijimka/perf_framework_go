# 🚀 Perf-Framework

Нагрузочный фреймворк на Go для тестирования производительности API. Поддерживает stateful-сценарии, многопоточность, различные профили нагрузки и генерацию HTML-отчетов.

##  Возможности

- ✅ **Многопоточность** — горутины Go для высокой производительности
- ✅ **Stateful-сценарии** — поддержка авторизации, токенов и последовательных шагов
- ✅ **Гибкая конфигурация** — YAML-конфиги для быстрого изменения параметров
- ✅ **Профили нагрузки** — Constant (постоянная) и Ramp-up (ступенчатая)
- ✅ **Сбор метрик** — RPS, время ответа, перцентили (P95), ошибки
- ✅ **HTML-отчеты** — автоматическая генерация красивых отчетов
- ✅ **Docker** — полная изоляция и воспроизводимость тестов
- ✅ **CI/CD** — готовый пайплайн для GitLab CI

## 📁 Структура проекта

```text
perf-framework/
├── cmd/
│   └── runner/
│       ├── main.go              # Точка входа
│       └── report.html.tmpl     # Шаблон HTML-отчета
── internal/
│   ├── client/
│   │   └── client.go            # HTTP клиент
│   ├── config/
│   │   ── config.go            # Работа с конфигурацией
│   └── metrics/
│       └── metrics.go           # Сбор и агрегация метрик
├── mock-server/
│   ├── main.go                  # Мок-сервер для тестирования
│   └── Dockerfile               # Dockerfile для мок-сервера
├── config.yaml                  # Конфигурация тестов
├── docker-compose.yml           # Docker Compose для полного стека
├── Dockerfile                   # Dockerfile для фреймворка
├── perf.ps1                     # Команды для Windows / PowerShell
├── perf.sh                      # Команды для macOS / Linux (sh)
└── .gitlab-ci.yml               # CI/CD пайплайн для GitLab
```

## 🛠 Требования

Для локального запуска вам понадобится:

- **Go** 1.21+ — [Скачать](https://go.dev/dl/)
- **Docker** и **Docker Compose** — [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- **Windows**: PowerShell 5.1+
- **macOS / Linux**: стандартный `sh`

Go нужен для локальных команд `build`, `run` и `test`. Для команд `docker-*` нужен только Docker с Compose. Make больше не требуется.

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone <URL_РЕПОЗИТОРИЯ>
cd perf-framework
```

### 2. Запуск мок-сервера

Откройте **первый терминал** и запустите мок-сервер (эмулирует тестируемый API):

Windows (PowerShell):

```powershell
.\perf.ps1 run
```

macOS / Linux:

```sh
./perf.sh run
```

Сервер запустится на `http://localhost:8080`.

### 3. Запуск тестов

Откройте **второй терминал** и запустите нагрузочные тесты:

Windows (PowerShell):

```powershell
.\perf.ps1 test
```

macOS / Linux:

```sh
./perf.sh test
```

Вы увидите вывод метрик в консоли и автоматически сгенерированный HTML-отчет в папке `reports/`.

### 4. Просмотр отчета

Откройте файл `reports/report_<дата>_<время>.html` в браузере.

---

## ⚙️ Конфигурация

Все параметры тестов задаются в файле `config.yaml`.

### Пример конфигурации

```yaml
target:
  base_url: "http://localhost:8080"  # URL тестируемого сервиса
  timeout_ms: 5000                    # Таймаут запроса в мс

# Профиль нагрузки
load_profile:
  type: "ramp-up"                    # constant или ramp-up
  ramp_up_duration_sec: 10           # Время нарастания (для ramp-up)
  target_workers: 10                 # Целевое количество воркеров

# Сценарии тестирования
scenarios:
  - name: "authorized_user_flow"
    iterations: 50                   # Сколько раз каждый воркер пройдет сценарий
    steps:
      - name: "login"
        method: "POST"
        path: "/api/login"
        save_token: true             # Сохранить токен из ответа

      - name: "get_data"
        method: "GET"
        path: "/api/data"
        use_token: true              # Использовать сохраненный токен
```

### Параметры конфигурации

| Параметр | Описание | Пример |
|----------|----------|--------|
| `target.base_url` | URL тестируемого сервиса | `http://localhost:8080` |
| `target.timeout_ms` | Таймаут HTTP-запроса | `5000` |
| `load_profile.type` | Тип нагрузки: `constant` или `ramp-up` | `ramp-up` |
| `load_profile.ramp_up_duration_sec` | Время нарастания нагрузки (сек) | `10` |
| `load_profile.target_workers` | Количество параллельных воркеров | `10` |
| `scenarios[].iterations` | Итераций на воркера | `50` |
| `steps[].save_token` | Сохранить токен из ответа | `true` |
| `steps[].use_token` | Добавить токен в заголовок | `true` |

---

## 📊 Профили нагрузки

### Constant (Постоянная нагрузка)

Все воркеры запускаются одновременно. Подходит для тестирования стабильной нагрузки.

```yaml
load_profile:
  type: "constant"
  target_workers: 10
```

### Ramp-up (Ступенчатая нагрузка)

Воркеры запускаются постепенно, создавая плавное нарастание нагрузки. Помогает найти точку деградации сервиса.

```yaml
load_profile:
  type: "ramp-up"
  ramp_up_duration_sec: 10
  target_workers: 10
```

В этом примере 10 воркеров будут запускаться в течение 10 секунд (по 1 воркеру в секунду).

---

## 🐳 Docker

### Запуск всего стека в Docker

Если вы не хотите устанавливать Go локально, можно запустить всё в Docker:

Windows (PowerShell):

```powershell
.\perf.ps1 docker-test
```

macOS / Linux:

```sh
./perf.sh docker-test
```

Эта команда:
1. Соберет Docker-образы для мок-сервера и фреймворка
2. Запустит мок-сервер
3. Дождется готовности мок-сервера и запустит тесты
4. Выведет метрики в консоль
5. Скопирует HTML-отчеты из контейнера в локальную папку `reports/`

Отчеты копируются через Docker API, поэтому сохраняются и при удаленном Docker daemon (например, в Docker-in-Docker). Ошибка копирования отчета дает ненулевой код завершения; если запуск уже завершился с ошибкой, сохраняется его исходный код.

После прогона контейнеры остаются остановленными для диагностики. Удалить их вместе с сетью:

```sh
docker compose down
```

`docker-run` запускает стек до нажатия Ctrl+C и копирует отчеты после его остановки.

### Ручная сборка образов

Windows (PowerShell):

```powershell
.\perf.ps1 docker-build
```

macOS / Linux:

```sh
./perf.sh docker-build
```

---

## 📈 HTML-отчеты

После каждого прогона тестов автоматически генерируется HTML-отчет в папке `reports/`.

Отчет содержит:
- Общее количество запросов
- Количество успешных запросов и ошибок
- RPS (запросов в секунду)
- Среднее, минимальное и максимальное время ответа
- 95-й перцентиль (P95)
- Общее время теста

---

## 🔄 CI/CD (GitLab CI)

Проект включает готовый пайплайн для GitLab CI (`.gitlab-ci.yml`).

### Этапы пайплайна

1. **build-images** — сборка Docker-образов
2. **performance-test** — запуск тестов (вручную)

### Запуск в GitLab CI

1. Запушьте код в GitLab-репозиторий
2. Перейдите в **CI/CD → Pipelines**
3. Найдите джобу `performance-test` и нажмите кнопку **Play** (▶️)
4. После завершения скачайте HTML-отчет из **Artifacts**

### Локальная проверка пайплайна

Команда `./perf.sh docker-test` (или `.\perf.ps1 docker-test` в PowerShell) запускает тесты так же, как CI/CD. Код завершения берется из контейнера `perf-test`; ошибка выгрузки отчета также завершает команду с ошибкой.

Текущий runner выводит HTTP-ошибки в метриках, но сам по себе не возвращает ненулевой код при таких ошибках. Порог ошибок/RPS/P95 для провала нагрузочного теста пока не задан.

---

### Jenkins

Проект включает `Jenkinsfile` с поддержкой:
- **Ручного запуска** с параметрами (профиль нагрузки, количество воркеров, итерации)
- **Автоматического запуска по таймеру** (каждые 5 минут)
- **Запуска по триггеру** от другой джобы (downstream/upstream)
- **Сохранения HTML-отчетов** как артефактов сборки
- **Умного управления кэшем Docker** (параметр `FORCE_DOCKER_REBUILD` для точечного сброса кэша)

#### Локальный запуск Jenkins

Для локального тестирования пайплайна можно поднять Jenkins в Docker:

```bash
# Запуск Jenkins
docker compose -f docker-compose-jenkins.yml up -d

# Получение начального пароля
docker exec jenkins cat /var/jenkins_home/secrets/initialAdminPassword
```

Откройте `http://localhost:8080`, введите пароль и завершите первоначальную настройку (установите рекомендуемые плагины).

**Необходимые плагины Jenkins:**
- **Docker Pipeline** — для работы с Docker в pipeline
- **HTML Publisher** — для просмотра HTML-отчетов в интерфейсе
- **Git** — для клонирования репозитория

#### Настройка Content Security Policy (CSP)

Для корректного отображения стилей и графиков Chart.js в отчетах Jenkins необходимо настроить CSP. Перейдите в **Manage Jenkins → Script Console** и выполните:

```groovy
System.setProperty("hudson.model.DirectoryBrowserSupport.CSP", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
```

#### Настройка джоб в Jenkins

**1. Ручной запуск тестов (`Perf-tests-manual`):**
- Создайте Pipeline джобу
- В секции **Pipeline** выберите **Pipeline script from SCM**
- Укажите ваш Git-репозиторий и путь `Jenkinsfile`
- Джоба будет доступна через **Build with Parameters**

**2. Автоматический запуск по таймеру:**
- Создайте вторую джобу (например, `Perf-tests-scheduled`)
- В настройках джобы включите **Build periodically**
- Укажите расписание: `H/5 * * * *` (каждые 5 минут)
- Используйте тот же `Jenkinsfile`

**3. Запуск по триггеру от другой джобы:**

Создайте upstream-джобу (например, `Unit-Tests`) с таким `Jenkinsfile`:

```groovy
pipeline {
    agent any
    
    stages {
        stage('Run Unit Tests') {
            steps {
                echo '🧪 Запускаем модульные тесты...'
                sh 'go test ./...'
            }
        }
    }
    
    post {
        success {
            echo '🚀 Тесты прошли успешно. Запускаем нагрузочные тесты в фоне...'
            build job: 'Perf-tests-manual', 
                 wait: false,  // Асинхронный запуск — не ждать завершения
                 parameters: [
                     string(name: 'LOAD_PROFILE', value: 'constant'),
                     string(name: 'TARGET_WORKERS', value: '10'),
                     string(name: 'ITERATIONS', value: '50')
                 ]
        }
    }
}
```

После успешного завершения `Unit-Tests` автоматически запустится `Perf-tests-manual` в фоне (параметр `wait: false` гарантирует, что upstream-джоба не будет ждать завершения нагрузочных тестов).

#### Управление кэшем Docker

При изменении файлов (например, шаблона отчета `report.html.tmpl`) Docker может использовать закэшированные слои. Для принудительной пересборки:

1. В `Jenkinsfile` добавлен параметр `FORCE_DOCKER_REBUILD` (галочка в интерфейсе Jenkins)
2. В `Dockerfile` объявлен `ARG CACHE_BUST=1` после `FROM`
3. В `docker-compose.yml` для сервиса `perf-test` передается `CACHE_BUST: ${CACHE_BUST:-1}`

Когда галочка `FORCE_DOCKER_REBUILD` установлена, Jenkins передает уникальный `BUILD_ID` в `CACHE_BUST`, что заставляет Docker пересобрать образ с нуля. По умолчанию (галочка снята) используется быстрый кэш.

#### Просмотр отчетов в Jenkins

1. После успешной сборки откройте джобу `Perf-tests-manual`
2. В левом меню найдите ссылку **Performance Report** (требует плагин HTML Publisher)
3. Или скачайте HTML-файл из раздела **Build Artifacts**

---

## 🛠 Расширение фреймворка

### Добавление нового сценария

Откройте `config.yaml` и добавьте новый сценарий:

```yaml
scenarios:
  - name: "new_scenario"
    iterations: 100
    steps:
      - name: "step_1"
        method: "POST"
        path: "/api/endpoint"
```

### Добавление нового клиента

Если вам нужно тестировать не HTTP API (например, gRPC или WebSocket), создайте новый клиент в папке `internal/client/`.

### Кастомные метрики

Откройте `internal/metrics/metrics.go` и добавьте новые поля в структуру `Metrics`.

---

## 📝 Команды скриптов

Без аргументов скрипты показывают справку. Все пути считаются от папки скрипта, поэтому его можно запускать из другого каталога по полному пути.

`build` создает `bin/perf-runner.exe` для Windows и `bin/perf-runner` для остальных ОС (по `go env GOOS`). `clean` удаляет папки `bin/` и `reports/`.

Если Windows блокирует запуск `.ps1` политикой выполнения, используйте разовый запуск:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\perf.ps1 help
```

В macOS / Linux также можно запускать скрипт через `sh perf.sh <команда>` без права на исполнение.

| Windows (PowerShell) | macOS / Linux | Описание |
|----------------------|---------------|----------|
| `.\perf.ps1 help` | `./perf.sh help` | Показать список доступных команд |
| `.\perf.ps1 build` | `./perf.sh build` | Собрать фреймворк локально |
| `.\perf.ps1 run` | `./perf.sh run` | Запустить мок-сервер |
| `.\perf.ps1 test` | `./perf.sh test` | Запустить тесты локально |
| `.\perf.ps1 clean` | `./perf.sh clean` | Очистить артефакты и отчеты |
| `.\perf.ps1 docker-build` | `./perf.sh docker-build` | Собрать Docker-образы |
| `.\perf.ps1 docker-run` | `./perf.sh docker-run` | Запустить весь стек в Docker |
| `.\perf.ps1 docker-test` | `./perf.sh docker-test` | Запустить тесты в Docker |

---
