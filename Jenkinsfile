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
        stage('Apply Migrations') {
            steps {
                script {
                    echo "Применяем миграции базы данных..."
                    withCredentials([
                        string(credentialsId: 'sop-postgres-user', variable: 'POSTGRES_USER'),
                        string(credentialsId: 'sop-postgres-password', variable: 'POSTGRES_PASSWORD'),
                        string(credentialsId: 'sop-postgres-db', variable: 'POSTGRES_DB'),
                        string(credentialsId: 'sop-postgres-host', variable: 'POSTGRES_HOST'),
                        string(credentialsId: 'sop-postgres-port', variable: 'POSTGRES_PORT')
                    ]) {
                        sh 'docker-compose run --rm migrator'
                    }
                }
            }
        }
        stage('Deploy Services') {
            steps {
                script {
                    echo "Миграции применены. Деплоим весь стек..."
                    withCredentials([
                        string(credentialsId: 'sop-postgres-user', variable: 'POSTGRES_USER'),
                        string(credentialsId: 'sop-postgres-password', variable: 'POSTGRES_PASSWORD'),
                        string(credentialsId: 'sop-postgres-db', variable: 'POSTGRES_DB'),
                        string(credentialsId: 'sop-postgres-host', variable: 'POSTGRES_HOST'),
                        string(credentialsId: 'sop-postgres-port', variable: 'POSTGRES_PORT'),
                        
                        string(credentialsId: 'sop-rabbitmq-user', variable: 'RABBITMQ_USER'),
                        string(credentialsId: 'sop-rabbitmq-pass', variable: 'RABBITMQ_PASS'),
                        string(credentialsId: 'sop-rabbitmq-host', variable: 'RABBITMQ_HOST'),
                        string(credentialsId: 'sop-rabbitmq-port', variable: 'RABBITMQ_PORT'),

                        string(credentialsId: 'sop-serv-amqp-queuename', variable: 'SERV_AMQP_QUEUENAME'),
                        string(credentialsId: 'sop-prov-amqp-queuename', variable: 'PROV_AMQP_QUEUENAME'),
                        string(credentialsId: 'sop-prov-app-provisioningtime', variable: 'PROV_APP_PROVISIONINGTIME'),

                        string(credentialsId: 'sop-grafana-user', variable: 'GRAFANA_USER'),
                        string(credentialsId: 'sop-grafana-password', variable: 'GRAFANA_PASSWORD')
                    ]) {
                        sh 'docker-compose up -d --build'
                    }
                }
            }
        }
        stage('Health Check') {
            steps {
                script {
                    sleep 15
                    sh 'curl -f http://hosting-api:8080/health || echo "Health check failed (hosting-api)"'
                    sh 'curl -f http://hosting-provisioner:7070/health || echo "Health check failed (hosting-provisioner)"'
                }
            }
        }
    }
}