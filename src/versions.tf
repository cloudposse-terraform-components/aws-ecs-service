terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.66.1, < 6.0.0"
    }
    template = {
      source  = "cloudposse/template"
      version = ">= 2.2"
    }
    jq = {
      source  = "massdriver-cloud/jq"
      version = ">=0.2.0"
    }
  }
}
