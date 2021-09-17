// seedjob.groovy

// create an array with our two pipelines
pipelines = ["first-pipeline", "another-pipeline"]

// iterate through the array and call the create_pipeline method
pipelines.each { pipeline ->
    println "Creating pipeline ${pipeline}"
    create_pipeline(pipeline)
}

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
                echo "parameter 'sValue' is: " + getProperty("sValue")
            }
        }
    }
}

                """)
            }
        }
        parameters {
            stringParam("sValue","NOP","value from the trigger")
        }
    }
}