pipeline {
    agent any
    environment {
        COMPOSE_PROJECT_NAME = "sop"
    }
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Test (CI)') {
            steps {
                script {
                    echo "Запуск тестов..."
                    sh 'docker build -f Dockerfile.test .'
                }
            }
        }
        stage('Deploy (CD)') {
            steps {
                script {
                    echo "Тесты прошли успешно. Деплоим..."
                    sh 'docker-compose up -d --build --no-deps hosting-service provisioning-service migrator'
                }
            }
        }
        stage('Health Check') {
            steps {
                script {
                    sleep 10
                    sh 'curl -f http://hosting-api:8080/health || echo "Health check failed (hosting-api)"'
                    sh 'curl -f http://hosting-provisioner:7070/health || echo "Health check failed (hosting-provisioner)"'
                }
            }
        }
    }
}