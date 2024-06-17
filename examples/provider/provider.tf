terraform {
  required_providers {
    dbsnapper = {
      source  = "dbsnapper/dbsnapper"
      version = "~>0.1"
    }
  }
}

provider "dbsnapper" {
  # Authentication token for API access
  # Omit this if you want to use DBSNAPPER_AUTHTOKEN environment variable
  authtoken = var.dbsnapper_authtoken
}