
file "deployment-pipeline.yaml" {
    dir = "./.codecatalyst/workflows"
    data = {
        Name          = "deployment-pipeline"
        RunMode       = "SUPERSEDED"
        SchemaVersion = "1.0"
        Triggers = [
            {
                Branches = [
                    "revamp", "main"
                ]
                Type = "PUSH"
            }
        ]
        Actions = {
            Build           = Actions.Build
            Beta            = Actions.Beta
            Gamma-us-east-1 = Actions.Gamma-us-east-1
            Gamma-us-west-2 = Actions.Gamma-us-west-2
            Prod-us-east-1  = Actions.Prod-us-east-1
            Prod-us-west-2  = Actions.Prod-us-west-2
        }
    }
}

Actions "Beta" {
    Actions "Deploy" {
        Configuration = {
            CdkRootPath = "infrastructure"
            Context     = "{\"deploymentConfigurationName\":\"CodeDeployDefault.ECSCanary10Percent5Minutes\"}"
            Region      = "us-east-1"
            StackName   = "sample-cdk-pipeline"
        }
        DependsOn = [
            "Build"
        ]
        Environment = {
            Connections = [
                {
                    Name = "main"
                    Role = "CodeCatalystWorkflowDevelopmentRole-nugg.xyz"
                }
            ]
            Name = "beta"
        }
        Identifier = "aws/cdk-deploy@v1"
        Inputs = {
            Sources = [
                "WorkflowSource"
            ]
        }
    }
    Actions "Test" {
        Configuration = {
            Steps = [
                {
                    Run = "mvn --batch-mode --no-transfer-progress soapui:test -Dsoapui.endpoint=$${endpointUrl}"
                }
                ,
                {
                    Run = <<EOT
					mvn --batch-mode --no-transfer-progress compile jmeter:jmeter jmeter:results -Djmeter.endpoint=$${endpointUrl} -Djmeter.threads=300 -Djmeter.duration=300 -Djmeter.throughput=6000
				EOT
                }
            ]
        }
        Identifier = "aws/managed-test@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
            Variables = [
                {
                    Name  = "endpointUrl"
                    Value = "$${Deploy.endpointUrl}"
                }
            ]
        }
        Outputs "AutoDiscoverReports" {
            Enabled = true
            IncludePaths = [
                "target/soapui-reports/*"
            ]
            ReportNamePrefix = "Beta"
            SuccessCriteria = {
                PassRate = 100
            }
        }
    }
    DependsOn = [
        "Build"
    ]
}

Actions "Build" {
    Actions "Package" {
        Compute = {
            Fleet = "Linux.Arm64.Large"
            Type  = "EC2"
        }
        Configuration = {
            Steps = [
                {
                    Run = <<EOT
						cd /root/.goenv/plugins/go-build/../.. && git pull && cd - && goenv install 1.21.5 && goenv global 1.21.5
					EOT
                }
                ,
                {
                    Run = "npm install -g @go-task/cli"
                }
                ,
                {
                    Run = "task test-ci"
                }
                ,
                {
                    Run = "go build -o ./bin/out ./cmd"
                }
            ]
        }
        Identifier = "aws/build@v1"
        Inputs = {
            Sources = [
                "WorkflowSource"
            ]
        }
        Outputs = {
            Artifacts = [
                {
                    Files = [
                        "./bin/out"
                    ]
                    Name = "package"
                }
            ]
            AutoDiscoverReports = {
                Enabled          = true
                ReportNamePrefix = "build"
                SuccessCriteria = {
                    PassRate = 100
                }
            }
        }
    }
    Actions "Synth" {
        Compute = {
            Fleet = "Linux.Arm64.Large"
            Type  = "EC2"
        }
        Configuration = {
            Steps = [
                {
                    Run = "cd /root/.goenv/plugins/go-build/../.. && git pull && cd - && goenv install 1.21.5 && goenv global 1.21.5"
                }
                ,
                {
                    Run = "npm install -g @go-task/cli"
                }
                ,
                {
                    Run = "task test-cdk-ci"
                }
                ,
                {
                    Run = "task cdk-synth"
                }
                ,
                {
                    Run = "mv infrastructure/cdk.out ."
                }
                ,
                {
                    Run = "mv infrastructure/cdk.json ."
                }
            ]
        }
        Identifier = "aws/build@v1"
        Inputs = {
            Sources = [
                "WorkflowSource"
            ]
        }
        Outputs = {
            Artifacts = [
                {
                    Files = [
                        "cdk.out/**/*", "cdk.json"
                    ]
                    Name = "synth"
                }
            ]
            AutoDiscoverReports = {
                Enabled = true
                IncludePaths = [
                    "test-reports/*"
                ]
                ReportNamePrefix = "synth"
                SuccessCriteria = {
                    PassRate = 100
                }
            }
        }
    }
}

Actions "Gamma-us-east-1" {
    Actions "Deploy" {
        Configuration = {
            CfnOutputVariables = "[\"endpointUrl\"]"
            Context            = "{\"deploymentConfigurationName\":\"CodeDeployDefault.ECSCanary10Percent5Minutes\"}"
            Region             = "us-east-1"
            StackName          = "fruit-api"
        }
        DependsOn = [
            "Beta"
        ]
        Environment = {
            Connections = [
                {
                    Name = "gamma"
                    Role = "codecatalyst"
                }
            ]
            Name = "Gamma"
        }
        Identifier = "aws/cdk-deploy@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
        }
    }
    Actions "Test" {
        Configuration = {
            Steps = [
                {
                    Run = "mvn --batch-mode --no-transfer-progress soapui:test -Dsoapui.endpoint=$${endpointUrl}"
                }
                ,
                {
                    Run = "mvn --batch-mode --no-transfer-progress compile jmeter:jmeter jmeter:results -Djmeter.endpoint=$${endpointUrl} -Djmeter.threads=300 -Djmeter.duration=300 -Djmeter.throughput=6000"
                }
            ]
        }
        Identifier = "aws/managed-test@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
            Variables = [
                {
                    Name  = "endpointUrl"
                    Value = "$${Deploy.endpointUrl}"
                }
            ]
        }
        Outputs "AutoDiscoverReports" {
            Enabled = true
            IncludePaths = [
                "target/soapui-reports/*"
            ]
            ReportNamePrefix = "Gamma-us-east-1"
            SuccessCriteria = {
                PassRate = 100
            }
        }
    }
    DependsOn = [
        "Beta"
    ]
}

Actions "Gamma-us-west-2" {
    Actions "Deploy" {
        Configuration = {
            CfnOutputVariables = "[\"endpointUrl\"]"
            Context            = "{\"deploymentConfigurationName\":\"CodeDeployDefault.ECSCanary10Percent5Minutes\"}"
            Region             = "us-west-2"
            StackName          = "fruit-api"
        }
        DependsOn = [
            "Beta"
        ]
        Environment = {
            Connections = [
                {
                    Name = "gamma"
                    Role = "codecatalyst"
                }
            ]
            Name = "Gamma"
        }
        Identifier = "aws/cdk-deploy@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
        }
    }
    Actions "Test" {
        Configuration = {
            Steps = [
                {
                    Run = "mvn --batch-mode --no-transfer-progress soapui:test -Dsoapui.endpoint=$${endpointUrl}"
                }
                ,
                {
                    Run = "mvn --batch-mode --no-transfer-progress compile jmeter:jmeter jmeter:results -Djmeter.endpoint=$${endpointUrl} -Djmeter.threads=300 -Djmeter.duration=300 -Djmeter.throughput=6000"
                }
            ]
        }
        Identifier = "aws/managed-test@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
            Variables = [
                {
                    Name  = "endpointUrl"
                    Value = "$${Deploy.endpointUrl}"
                }
            ]
        }
        Outputs "AutoDiscoverReports" {
            Enabled = true
            IncludePaths = [
                "target/soapui-reports/*"
            ]
            ReportNamePrefix = "Gamma-us-west-2"
            SuccessCriteria = {
                PassRate = 100
            }
        }
    }
    DependsOn = [
        "Beta"
    ]
}

Actions "Prod-us-east-1" {
    Actions "Deploy" {
        Configuration = {
            CfnOutputVariables = "[\"endpointUrl\"]"
            Context            = "{\"deploymentConfigurationName\":\"CodeDeployDefault.ECSCanary10Percent5Minutes\"}"
            Region             = "us-east-1"
            StackName          = "fruit-api"
        }
        DependsOn = [
            "Gamma-us-west-2", "Gamma-us-east-1"
        ]
        Environment = {
            Connections = [
                {
                    Name = "prod"
                    Role = "codecatalyst"
                }
            ]
            Name = "Production"
        }
        Identifier = "aws/cdk-deploy@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
        }
    }
    DependsOn = [
        "Gamma-us-west-2", "Gamma-us-east-1"
    ]
}

Actions "Prod-us-west-2" {
    Actions "Deploy" {
        Configuration = {
            CfnOutputVariables = "[\"endpointUrl\"]"
            Context            = "{\"deploymentConfigurationName\":\"CodeDeployDefault.ECSCanary10Percent5Minutes\"}"
            Region             = "us-west-2"
            StackName          = "fruit-api"
        }
        DependsOn = [
            "Gamma-us-west-2", "Gamma-us-east-1"
        ]
        Environment = {
            Connections = [
                {
                    Name = "prod"
                    Role = "codecatalyst"
                }
            ]
            Name = "Production"
        }
        Identifier = "aws/cdk-deploy@v1"
        Inputs = {
            Artifacts = [
                "synth"
            ]
        }
    }
    DependsOn = [
        "Gamma-us-west-2", "Gamma-us-east-1"
    ]
}
