properties([buildDiscarder(logRotator(numToKeepStr: '20'))])

pipeline {
    agent {
        label 'linux'
    }

    options {
        skipDefaultCheckout(true)
    }

    stages {
        stage('Build') {
            agent {
                label 'linux'
            }
            steps  {
                dir('src/github.com/docker/lunchbox') {
                    script {
                        try {
                            checkout scm
                            sh 'rm -rf *.tar.gz stash'
                            sh 'docker image prune -f'
                            sh 'make ci-lint'
                            sh 'make ci-test'
                            sh 'make ci-bin-all'
                            sh 'mkdir stash'
                            sh 'ls *.tar.gz | xargs -i tar -xf {} -C stash'
                            dir('stash') {
                                stash name: 'e2e'
                            }
                            if(!(env.BRANCH_NAME ==~ "PR-\\d+")) {
                                archiveArtifacts '*.tar.gz'
                            }
                        } finally {
                            def clean_images = /docker image ls --format "{{.ID}}\t{{.Tag}}" | grep $(git describe --always --dirty) | awk '{print $1}' | xargs docker image rm/
                            sh clean_images
                        }
                    }
                }
            }
        }
        stage('Test') {
            parallel {
                stage("Test Linux") {
                    agent {
                        label 'linux'
                    }
                    steps  {
                        dir('src/github.com/docker/lunchbox') {
                            deleteDir()
                            unstash 'e2e'
                            sh './docker-app-e2e-linux'
                        }
                    }
                }
                stage("Test Mac") {
                    agent {
                        label "mac"
                    }
                    steps {
                        dir('src/github.com/docker/lunchbox') {
                            deleteDir()
                            unstash 'e2e'
                            sh './docker-app-e2e-darwin'
                        }
                    }
                }
                stage("Test Win") {
                    agent {
                        label "windows"
                    }
                    steps {
                        dir('src/github.com/docker/lunchbox') {
                            deleteDir()
                            unstash "e2e"
                            bat 'docker-app-e2e-windows.exe'
                        }
                    }
                }
            }
        }
    }
}
