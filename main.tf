terraform {
  required_version = "> 0.12"
}

### Variables ###
variable "project_id" {
  default = "proj-cloud-trace"
}

variable "region" {
  description = "Google Cloud Region (e.g. us-central1, see gcloud compute regions list)"
  default = "europe-west3"
}

### Google Provider configuration ###
provider "google" {
  project = var.project_id
  region = var.region
  version = "3.15"
}

### Google APIs ###
resource "google_project_service" "appengine_googleapis_com" {
  project = var.project_id
  service = "appengine.googleapis.com"
}

resource "google_project_service" "pubsub_googleapis_com" {
  project = var.project_id
  service = "pubsub.googleapis.com"
}

resource "google_project_service" "cloudfunctions_googleapis_com" {
  project = var.project_id
  service = "cloudfunctions.googleapis.com"
}

### App Engines ###
resource "google_app_engine_application" "trace_example" {
  depends_on = [
    google_project_service.appengine_googleapis_com,
  ]

  project = var.project_id
  location_id = var.region
}

resource "null_resource" "user_service" {
  depends_on = [
    google_app_engine_application.trace_example,
  ]

  provisioner "local-exec" {
    command = "gcloud app deploy ./user-service/app.yaml -q --promote"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "gcloud app services delete user-service"
  }
}

resource "null_resource" "account_service" {
  depends_on = [
    google_app_engine_application.trace_example,
  ]

  provisioner "local-exec" {
    command = "gcloud app deploy ./account-service/app.yaml -q --promote"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "gcloud app services delete account-service"
  }
}

### Pub/Sub ###
resource "google_pubsub_topic" "topic_send_mail" {
  depends_on = [
    google_project_service.pubsub_googleapis_com
  ]

  name    = "send-mail"
  project = var.project_id
}

resource "google_pubsub_subscription" "subscription_send_mail" {
  depends_on = [
    google_pubsub_topic.topic_send_mail
  ]

  name    = "send-mail-subscription"
  topic = "send-mail"

  push_config {
    push_endpoint = "https://${var.region}-${var.project_id}.cloudfunctions.net/send-mail"

    attributes = {
      x-goog-version = "v1"
    }

    oidc_token {
      service_account_email = "${var.project_id}@appspot.gserviceaccount.com"
    }
  }

  project = var.project_id
}

### Cloud Function ###
resource "null_resource" "send_mail" {
  depends_on = [
    google_pubsub_topic.topic_send_mail,
  ]

  provisioner "local-exec" {
    command = "gcloud functions deploy send-mail --entry-point=SendMail --source=. --ingress-settings=all --trigger-http --runtime go113 --region europe-west3 -q"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "gcloud functions delete send-mail"
  }
}