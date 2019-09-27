
terraform {
  required_providers {
    provider "aws" {
      version = "~> 1.0.0"
      source  = "hashicorp/aws"
    }
    provider "consul" {
      source  = "tf.example.com/hashicorp/consul"
      version = "~> 1.2.0"
    }
  }
}
