terraform {
  required_providers {
    commonfatedeployment = {
      source  = "common-fate/deploymeta"
      version = "0.1.0"
    }
  }
}


provider "deploymeta" {
  # example configuration here
  deployment_name = "dev"
  licence_key     = "XXX"
}

data "deploymeta_deployment" "this" {}

output "deployment_id" {
  value = data.deploymeta_deployment.this.id
}

# resource "commonfatedeployment_monitoring_write_token" "main" {}
