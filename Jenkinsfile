pipeline {
    agent any

    // Параметры для ручного запуска
    parameters {
        choice(
            name: 'LOAD_PROFILE',
            choices: ['constant', 'ramp-up'],
            description: 'Выберите профиль нагрузки'
        )
        string(
            name: 'TARGET_WORKERS',
            defaultValue: '10',
            description: 'Количество воркеров'
        )
        string(
            name: 'ITERATIONS',
            defaultValue: '50',
            description: 'Количество итераций на воркера'
        )
        booleanParam(
            name: 'GENERATE_REPORT',
            defaultValue: true,
            description: 'Генерировать HTML-отчет'
        )
    }

    options {
        // Даем 5 секунд "тишины" между триггерами, чтобы ручной клик и таймер не создали 2 джобы одновременно
        quietPeriod(5)
        // Запрещаем запускать эту джобу параллельно саму с собой (защита от наложений)
        disableConcurrentBuilds()
    }

    environment {
        DOCKER_COMPOSE = 'docker compose'
        REPORT_DIR = 'reports'
    }

    stages {
        stage('Cleanup Previous Runs') {
            steps {
                echo '🧹 Очищаем возможные zombie-контейнеры от предыдущих запусков...'
                sh 'docker compose down --remove-orphans || true'
            }
        }

        stage('Checkout') {
            steps {
                echo '📥 Клонируем репозиторий...'
                checkout scm
            }
        }

        stage('Build Docker Images') {
            steps {
                echo ' Собираем Docker образы...'
                sh '${DOCKER_COMPOSE} build'
            }
        }

        stage('Prepare Config') {
            steps {
                echo '⚙️ Готовим конфигурацию...'
                script {
                    def configFile = 'config.yaml'
                    if (fileExists(configFile)) {
                        def content = readFile(configFile)
                        
                        // Используем параметры или безопасные значения по умолчанию
                        def profile = params.LOAD_PROFILE ?: 'constant'
                        def workers = params.TARGET_WORKERS ?: '10'
                        def iterations = params.ITERATIONS ?: '50'
                        
                        content = content.replaceAll(/type:\s*["']?\w+["']?/, "type: \"${profile}\"")
                        content = content.replaceAll(/target_workers:\s*\d+/, "target_workers: ${workers}")
                        content = content.replaceAll(/iterations:\s*\d+/, "iterations: ${iterations}")
                        
                        writeFile file: configFile, text: content
                        echo "✅ Конфигурация обновлена: профиль=${profile}, воркеров=${workers}, итераций=${iterations}"
                    } else {
                        echo "⚠️ Файл config.yaml не найден, используем дефолтный."
                    }
                }
            }
        }

        stage('Prepare Reports Directory') {
            steps {
                echo '📁 Создаем папку для отчетов...'
                script {
                    // Просто создаем, если нет. Никаких удалений!
                    if (!fileExists(REPORT_DIR)) {
                        sh "mkdir -p ${REPORT_DIR}"
                    }
                    echo "✅ Папка ${REPORT_DIR} готова"
                }
            }
        }

        stage('Run Performance Tests') {
            steps {
                echo ' Запускаем нагрузочные тесты...'
                sh '${DOCKER_COMPOSE} up --build --abort-on-container-exit perf-test'
            }
        }

        stage('Debug Reports on Host') {
            steps {
                echo '🔍 Проверяем, попал ли файл из контейнера на хост Jenkins...'
                sh 'echo "Текущая директория: $(pwd)"'
                sh 'ls -la reports/ || echo "Папка reports пуста или не существует!"'
                sh 'find . -name "*.html" || echo "HTML файлы не найдены"'
            }
        }

                stage('Extract Reports from Container') {
            steps {
                echo '📦 Извлекаем отчет из контейнера через docker cp...'
                script {
                    def containerId = sh(script: 'docker compose ps -a -q perf-test', returnStdout: true).trim()
                    
                    if (containerId) {
                        echo "✅ Найден контейнер: ${containerId}"
                        
                        echo '️  Очищаем папку reports от старых файлов...'
                        sh "rm -f ${REPORT_DIR}/*.html || true"
                        // =================================================
                        
                        sh "mkdir -p ${REPORT_DIR}"
                        sh "docker cp ${containerId}:/app/reports/. ./${REPORT_DIR}/"
                        sh "chmod -R 777 ${REPORT_DIR} || true"
                        
                        echo "🔍 Проверка содержимого:"
                        sh "ls -la ${REPORT_DIR}/"
                    } else {
                        echo "⚠️ Контейнер perf-test не найден."
                    }
                }
            }
        }

        stage('Collect Reports') {
            when {
                expression { params.GENERATE_REPORT == true }
            }
            steps {
                echo '📊 Собираем отчеты...'
                script {
                    // Проверяем, есть ли отчеты
                    def reportExists = fileExists("${REPORT_DIR}")
                    if (reportExists) {
                        echo "✅ Отчеты найдены в папке ${REPORT_DIR}"
                    } else {
                        echo "⚠️  Папка с отчетами не найдена"
                    }
                }
            }
        }
    }

        post {
        always {
            script {
                echo '📦 Архивируем отчеты...'
                
                // "'reports/**/*.html' doesn't match anything, but '**/*.html' does"
                archiveArtifacts artifacts: '**/*.html', allowEmptyArchive: true
                
                publishHTML([
                    allowMissing: true,
                    alwaysLinkToLastBuild: true,
                    keepAll: true, 
                    reportDir: "${REPORT_DIR}",
                    reportFiles: '*.html',
                    reportName: 'Performance Report',
                    reportTitles: 'Нагрузочное тестирование'
                ])
            }
            
            echo '🧹 Очистка после завершения...'
            sh '${DOCKER_COMPOSE} down || true'
        }
        
        success {
            echo '✅ Тесты успешно завершены!'
        }
        
        failure {
            echo '❌ Тесты завершились с ошибками!'
        }
    }
}