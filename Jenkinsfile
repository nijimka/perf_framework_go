pipeline {
    agent any

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
        quietPeriod(5)
        disableConcurrentBuilds()
    }

    environment {
        DOCKER_COMPOSE = 'docker compose'
        REPORT_DIR = 'reports'
    }

    stages {
        stage('Cleanup Previous Runs') {
            steps {
                echo '🧹 Очищаем возможные zombie-контейнеры...'
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
                echo '🐳 Собираем Docker образы...'
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
                        def profile = params.LOAD_PROFILE ?: 'constant'
                        def workers = params.TARGET_WORKERS ?: '10'
                        def iterations = params.ITERATIONS ?: '50'
                        
                        content = content.replaceAll(/type:\s*["']?\w+["']?/, "type: \"${profile}\"")
                        content = content.replaceAll(/target_workers:\s*\d+/, "target_workers: ${workers}")
                        content = content.replaceAll(/iterations:\s*\d+/, "iterations: ${iterations}")
                        
                        writeFile file: configFile, text: content
                        echo "✅ Конфигурация обновлена: профиль=${profile}, воркеров=${workers}, итераций=${iterations}"
                    }
                }
            }
        }

        stage('Run Performance Tests') {
            steps {
                echo '🚀 Запускаем нагрузочные тесты...'
                sh '${DOCKER_COMPOSE} up --build --abort-on-container-exit perf-test'
            }
        }

        stage('Extract Reports from Container') {
            steps {
                echo '📦 Извлекаем отчет из контейнера через docker cp...'
                script {
                    def containerId = sh(script: 'docker compose ps -a -q perf-test', returnStdout: true).trim()
                    
                    if (containerId) {
                        echo "✅ Найден контейнер: ${containerId}"
                        
                        // === ЯДЕРНАЯ ОЧИСТКА ===
                        echo '🗑️ ПОЛНОСТЬЮ удаляем папку reports, чтобы гарантировать чистоту...'
                        sh "rm -rf ${REPORT_DIR}"
                        sh "mkdir -p ${REPORT_DIR}"
                        // ========================
                        
                        // Копируем содержимое (точка в конце пути важна!)
                        sh "docker cp ${containerId}:/app/reports/. ./${REPORT_DIR}/"
                        sh "chmod -R 777 ${REPORT_DIR} || true"
                        
                        echo "🔍 ПРОВЕРКА: в папке должен быть ровно 1 файл:"
                        sh "ls -la ${REPORT_DIR}/"
                        sh "echo 'Количество HTML файлов:' && find ${REPORT_DIR} -maxdepth 1 -name '*.html' | wc -l"
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
                echo '📊 Финальная проверка перед архивацией...'
                script {
                    if (fileExists("${REPORT_DIR}")) {
                        echo "✅ Папка ${REPORT_DIR} существует"
                    }
                }
            }
        }
    }

    post {
        always {
            script {
                echo '📦 Архивируем отчеты...'
                
                // ИСПРАВЛЕНИЕ: ищем файлы СТРОГО в папке reports, а не во всем проекте
                archiveArtifacts artifacts: 'reports/*.html', allowEmptyArchive: true
                
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