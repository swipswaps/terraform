
terraform {
  required_providers {
    aws = {
      version = "~> 1.0.0"
      source  = "hashicorp/aws"
    }
    consul = {
      source  = "tf.example.com/hashicorp/consul"
      version = "~> 1.2.0"
    }
  }
}
