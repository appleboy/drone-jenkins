// seedjob.groovy

// create two pipelines
println "Creating two pipelines"
create_pipeline("first-pipeline")
create_pipeline2("another-pipeline")

// a method that creates a basic pipeline with the given parameter name
def create_pipeline(String name) {
    pipelineJob(name) {
        definition {
            cps {
                sandbox(false)
                script("""

// this is an example declarative pipeline that says hello and goodbye
pipeline {
    agent any
    stages {
        stage("Hello") {
            steps {
                echo "Hello from pipeline ${name}"
            }
        }
        stage("Goodbye") {
            steps {
                echo "Goodbye from pipeline ${name}"
            }
        }
    }
}

                """)
            }
        }
    }
}


// a method that creates a basic pipeline with the given parameter name
def create_pipeline2(String name) {
    pipelineJob(name) {
        definition {
            cps {
                sandbox(false)
                script("""

// this is an example declarative pipeline that says hello and goodbye
pipeline {
    agent any
    stages {
        stage("Hello") {
            steps {
                echo "Hello from pipeline ${name}"
            }
        }
        stage("Goodbye") {
            steps {
                echo "Goodbye from pipeline ${name}"
                echo "parameter 'sValue' is: " + getProperty("sValue")
                echo "parameter 'sValue2' is: " + getProperty("sValue2")
            }
        }
    }
}

                """)
            }
        }
        parameters {
            stringParam("sValue","NOP","value from the trigger")
            stringParam("sValue2","NOP","value from the trigger")
        }
    }
}